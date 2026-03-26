package ffmpeg_go

import (
	"fmt"
	"slices"
	"strings"
)

func getInputArgs(node *Node) []string {
	var args []string
	if node.name == "input" {
		kwargs := node.kwargs.Copy()
		filename := kwargs.PopString("filename")
		format := kwargs.PopString("format")
		videoSize := kwargs.PopString("video_size")
		if format != "" {
			args = append(args, "-f", format)
		}
		if videoSize != "" {
			args = append(args, "-video_size", videoSize)
		}
		args = append(args, ConvertKwargsToCmdLineArgs(kwargs)...)
		args = append(args, "-i", filename)
	} else {
		panic("unsupported node input name")
	}
	return args
}

func formatInputStreamName(streamNameMap map[string]string, edge DagEdge, finalArg bool) string {
	prefix := streamNameMap[fmt.Sprintf("%d%s", edge.UpStreamNode.Hash(), edge.UpStreamLabel)]
	suffix := ""
	format := "[%s%s]"
	if edge.UpStreamSelector != "" {
		suffix = fmt.Sprintf(":%s", edge.UpStreamSelector)
	}
	if finalArg && edge.UpStreamNode.(*Node).nodeType == "InputNode" {
		format = "%s%s"
	}
	return fmt.Sprintf(format, prefix, suffix)
}

func formatOutStreamName(streamNameMap map[string]string, edge DagEdge) string {
	return fmt.Sprintf("[%s]", streamNameMap[fmt.Sprintf("%d%s", edge.UpStreamNode.Hash(), edge.UpStreamLabel)])
}

func _getFilterSpec(node *Node, outOutingEdgeMap map[Label][]NodeInfo, streamNameMap map[string]string) string {
	var input, output []string
	for _, e := range node.GetInComingEdges() {
		input = append(input, formatInputStreamName(streamNameMap, e, false))
	}
	outEdges := GetOutGoingEdges(node, outOutingEdgeMap)
	for _, e := range outEdges {
		output = append(output, formatOutStreamName(streamNameMap, e))
	}
	return fmt.Sprintf("%s%s%s", strings.Join(input, ""), node.GetFilter(outEdges), strings.Join(output, ""))
}

func _getAllLabelsInSorted(m map[Label]NodeInfo) []Label {
	var r []Label
	for a := range m {
		r = append(r, a)
	}
	slices.Sort(r)
	return r
}

func _getAllLabelsSorted(m map[Label][]NodeInfo) []Label {
	var r []Label
	for a := range m {
		r = append(r, a)
	}
	slices.Sort(r)
	return r
}

func _allocateFilterStreamNames(nodes []*Node, outOutingEdgeMaps map[int]map[Label][]NodeInfo, streamNameMap map[string]string) {
	sc := 0
	for _, n := range nodes {
		om := outOutingEdgeMaps[n.Hash()]
		// todo sort
		for _, l := range _getAllLabelsSorted(om) {
			if len(om[l]) > 1 {
				panic(fmt.Sprintf(`encountered %s with multiple outgoing edges
with same upstream label %s; a 'split'' filter is probably required`, n.name, l))
			}
			streamNameMap[fmt.Sprintf("%d%s", n.Hash(), l)] = fmt.Sprintf("s%d", sc)
			sc += 1
		}
	}
}

func _getFilterArg(nodes []*Node, outOutingEdgeMaps map[int]map[Label][]NodeInfo, streamNameMap map[string]string) string {
	_allocateFilterStreamNames(nodes, outOutingEdgeMaps, streamNameMap)
	var filterSpec []string
	for _, n := range nodes {
		filterSpec = append(filterSpec, _getFilterSpec(n, outOutingEdgeMaps[n.Hash()], streamNameMap))
	}
	return strings.Join(filterSpec, ";")
}

func _getGlobalArgs(node *Node) []string {
	return node.args
}

// Recursively collects arguments only from the RawArgsNode chain.
// Due to the structure of the DAG, each output stream will only see its own branch.
func _collectRawArgs(node *Node) []string {
	var res []string
	for _, e := range node.GetInComingEdges() {
		if upNode, ok := e.UpStreamNode.(*Node); ok && upNode.nodeType == "RawArgsNode" {
			res = append(res, _collectRawArgs(upNode)...)
		}
	}
	res = append(res, node.args...)
	return res
}

func _getOutputArgs(node *Node, streamNameMap map[string]string) []string {
	if node.name != "output" {
		panic("Unsupported output node")
	}
	var args []string
	incomingEdges := node.GetInComingEdges()
	if len(incomingEdges) == 0 {
		panic("Output node has no mapped streams")
	}

	// -map_chapters
	var hasChapters bool
	for _, e := range incomingEdges {
		if e.UpStreamNode.(*Node).nodeType == "MapChaptersNode" {
			if hasChapters {
				panic(fmt.Sprintf("output '%s' has multiple map_chapters sources", node.kwargs["filename"]))
			}
			chaptersInputEdges := e.UpStreamNode.(*Node).GetInComingEdges()
			inputIdx := streamNameMap[fmt.Sprintf("%d", chaptersInputEdges[0].UpStreamNode.Hash())]

			args = append(args, "-map_chapters", inputIdx)
			hasChapters = true
		}
	}

	// -map
	for _, e := range incomingEdges {
		upNode := e.UpStreamNode.(*Node)
		// Inappropriate nodes are skipped.
		if upNode.nodeType == "MapChaptersNode" || upNode.nodeType == "RawArgsNode" || upNode.nodeType == "MapMetadataNode" {
			continue
		}

		streamName := formatInputStreamName(streamNameMap, e, true)
		if streamName != "0" || len(incomingEdges) > 1 {
			args = append(args, "-map", streamName)
		}
	}

	// -map_metadata after -map
	for _, e := range incomingEdges {
		if upNode := e.UpStreamNode.(*Node); upNode.nodeType == "MapMetadataNode" {
			metaInEdges := upNode.GetInComingEdges()
			inputIdx := streamNameMap[fmt.Sprintf("%d", metaInEdges[0].UpStreamNode.Hash())]

			ot := upNode.kwargs.GetString("ot")
			it := upNode.kwargs.GetString("it")

			flag := "-map_metadata"
			if ot != "" {
				flag = fmt.Sprintf("%s:%s", flag, ot)
			}

			argVal := inputIdx
			if it != "" {
				argVal = fmt.Sprintf("%s:%s", argVal, it)
			}

			args = append(args, flag, argVal)
		}
	}

	// RawArgs before Output
	for _, e := range incomingEdges {
		if upNode := e.UpStreamNode.(*Node); upNode.nodeType == "RawArgsNode" {
			args = append(args, _collectRawArgs(upNode)...)
		}
	}

	kwargs := node.kwargs.Copy()

	filename := kwargs.PopString("filename")
	if kwargs.HasKey("format") {
		args = append(args, "-f", kwargs.PopString("format"))
	}
	if kwargs.HasKey("video_bitrate") {
		args = append(args, "-b:v", kwargs.PopString("video_bitrate"))
	}
	if kwargs.HasKey("audio_bitrate") {
		args = append(args, "-b:a", kwargs.PopString("audio_bitrate"))
	}
	if kwargs.HasKey("video_size") {
		args = append(args, "-video_size", kwargs.PopString("video_size"))
	}

	args = append(args, ConvertKwargsToCmdLineArgs(kwargs)...)
	args = append(args, filename)
	return args
}

func (s *Stream) GetArgs() []string {
	var args []string
	nodes := getStreamSpecNodes([]*Stream{s})
	var dagNodes []DagNode
	streamNameMap := map[string]string{}
	for i := range nodes {
		dagNodes = append(dagNodes, nodes[i])
	}
	sorted, outGoingMap, err := TopSort(dagNodes)
	if err != nil {
		panic(err)
	}
	DebugNodes(sorted)
	DebugOutGoingMap(sorted, outGoingMap)
	var inputNodes, outputNodes, globalNodes, filterNodes []*Node
	for i := range sorted {
		n := sorted[i].(*Node)
		switch n.nodeType {
		case "InputNode":
			streamNameMap[fmt.Sprintf("%d", n.Hash())] = fmt.Sprintf("%d", len(inputNodes))
			inputNodes = append(inputNodes, n)
		case "OutputNode":
			outputNodes = append(outputNodes, n)
		case "GlobalNode":
			globalNodes = append(globalNodes, n)
		case "FilterNode":
			filterNodes = append(filterNodes, n)
		case "MapChaptersNode":
		case "MapMetadataNode":
		case "RawArgsNode":
			// If a "default: with panic()" section appears later,
			// unspecified nodes will end up there. To prevent them
			// from ending up there, all node types are listed.
		}
	}
	// input args from inputNodes
	for _, n := range inputNodes {
		args = append(args, getInputArgs(n)...)
	}
	// filter args from filterNodes
	filterArgs := _getFilterArg(filterNodes, outGoingMap, streamNameMap)
	if filterArgs != "" {
		args = append(args, "-filter_complex", filterArgs)
	}
	// output args from outputNodes
	for _, n := range outputNodes {
		args = append(args, _getOutputArgs(n, streamNameMap)...)
	}
	// global args with outputNodes
	for _, n := range globalNodes {
		args = append(args, _getGlobalArgs(n)...)
	}
	//
	return args
}
