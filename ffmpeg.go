package ffmpeg_go

import (
	"errors"
	"log"
)

// Input file URL (ffmpeg “-i“ option)
//
// Any supplied kwargs are passed to ffmpeg verbatim (e.g. “t=20“,
// “f='mp4'“, “acodec='pcm'“, etc.).
//
// To tell ffmpeg to read from stdin, use “pipe:“ as the filename.
//
// Official documentation: `Main options <https://ffmpeg.org/ffmpeg.html#Main-options>`__
func Input(filename string, kwargs ...KwArgs) *Stream {
	args := MergeKwArgs(kwargs)
	args["filename"] = filename
	if fmt := args.PopString("f"); fmt != "" {
		if args.HasKey("format") {
			panic(errors.New("can't specify both `format` and `f` options"))
		}
		args["format"] = fmt
	}
	return NewInputNode("input", nil, args).Stream("", "")
}

// Add extra global command-line argument(s), e.g. “-progress“.
func (s *Stream) GlobalArgs(args ...string) *Stream {
	if s.Type != "OutputStream" {
		panic("cannot overwrite outputs on non-OutputStream")
	}
	return NewGlobalNode("global_args", []*Stream{s}, args, nil).Stream("", "")
}

// Overwrite output files without asking (ffmpeg “-y“ option)
//
// Official documentation: `Main options <https://ffmpeg.org/ffmpeg.html#Main-options>`_
func (s *Stream) OverwriteOutput(stream *Stream) *Stream {
	if s.Type != "OutputStream" {
		panic("cannot overwrite outputs on non-OutputStream")
	}
	return NewGlobalNode("overwrite_output", []*Stream{stream}, []string{"-y"}, nil).Stream("", "")
}

// Include all given outputs in one ffmpeg command line
func MergeOutputs(streams ...*Stream) *Stream {
	return NewMergeOutputsNode("merge_output", streams).Stream("", "")
}

// Output file URL
//
//	Syntax:
//	    `ffmpeg.Output([]*Stream{stream1, stream2, stream3...}, filename, kwargs)`
//
//	Any supplied keyword arguments are passed to ffmpeg verbatim (e.g.
//	``t=20``, ``f='mp4'``, ``acodec='pcm'``, ``vcodec='rawvideo'``,
//	etc.).  Some keyword-arguments are handled specially, as shown below.
//
//	Args:
//	    video_bitrate: parameter for ``-b:v``, e.g. ``video_bitrate=1000``.
//	    audio_bitrate: parameter for ``-b:a``, e.g. ``audio_bitrate=200``.
//	    format: alias for ``-f`` parameter, e.g. ``format='mp4'``
//	        (equivalent to ``f='mp4'``).
//
//	If multiple streams are provided, they are mapped to the same
//	output.
//
//	To tell ffmpeg to write to stdout, use ``pipe:`` as the filename.
//
//	Official documentation: `Synopsis <https://ffmpeg.org/ffmpeg.html#Synopsis>`__
//	"""
func Output(streams []*Stream, fileName string, kwargs ...KwArgs) *Stream {
	args := MergeKwArgs(kwargs)
	if !args.HasKey("filename") {
		if fileName == "" {
			panic("filename must be provided")
		}
		args["filename"] = fileName
	}

	return NewOutputNode("output", streams, nil, args).Stream("", "")
}

// Output file URL
//
//	Syntax:
//	    `ffmpeg.Output(ctx, []*Stream{stream1, stream2, stream3...}, filename, kwargs)`
//
//	Any supplied keyword arguments are passed to ffmpeg verbatim (e.g.
//	``t=20``, ``f='mp4'``, ``acodec='pcm'``, ``vcodec='rawvideo'``,
//	etc.).  Some keyword-arguments are handled specially, as shown below.
//
//	Args:
//	    video_bitrate: parameter for ``-b:v``, e.g. ``video_bitrate=1000``.
//	    audio_bitrate: parameter for ``-b:a``, e.g. ``audio_bitrate=200``.
//	    format: alias for ``-f`` parameter, e.g. ``format='mp4'``
//	        (equivalent to ``f='mp4'``).
//
//	If multiple streams are provided, they are mapped to the same
//	output.
//
//	To tell ffmpeg to write to stdout, use ``pipe:`` as the filename.
//
//	Official documentation: `Synopsis <https://ffmpeg.org/ffmpeg.html#Synopsis>`__
//	"""
func (s *Stream) Output(fileName string, kwargs ...KwArgs) *Stream {
	if s.Type != "FilterableStream" && s.Type != "RawArgsStream" {
		log.Panic("cannot output on non-FilterableStream or non-RawArgsStream")
	}
	output := Output([]*Stream{s}, fileName, kwargs...)
	return output
}

// Method that initiates the creation of a MapChapters node from an InputNode node.
func (s *Stream) MapChapters() *Stream {
	AssertType(s.Type, "FilterableStream", "map_chapters")

	if s.Node.nodeType != "InputNode" {
		panic("MapChapters can only be called directly from an Input")
	}

	return NewMapChaptersNode("map_chapters", []*Stream{s}).Stream("", "")
}

// RawArgs creates an initial stream of raw command-line arguments.
func RawArgs(args ...string) *Stream {
	return NewRawArgsNode("raw_args", args, nil).Stream("", "")
}

// The RawArgs method adds new arguments to an existing RawArgsStream chain.ы
func (s *Stream) RawArgs(args ...string) *Stream {
	AssertType(s.Type, "RawArgsStream", "raw_args")
	return NewRawArgsNode("raw_args", args, []*Stream{s}).Stream("", "")
}

// it - from (Input Target), ot - to (Output Target)
func (s *Stream) MapMetadata(ot, it string) *Stream {
	AssertType(s.Type, "FilterableStream", "map_metadata")

	if s.Node.nodeType != "InputNode" {
		panic("MapMetadata can only be called directly from an Input")
	}

	kwargs := KwArgs{}
	if ot != "" {
		kwargs["ot"] = ot
	}
	if it != "" {
		kwargs["it"] = it
	}

	return NewMapMetadataNode("map_metadata", []*Stream{s}, kwargs).Stream("", "")
}
