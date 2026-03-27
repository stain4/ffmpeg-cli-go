# ffmpeg-cli-go

ffmpeg-cli-go is based on https://github.com/u2takey/ffmpeg-go. It has been significantly modified and refactored.

## Key Differences from ffmpeg-go

This project is a streamlined fork of [u2takey/ffmpeg-go](https://github.com/u2takey/ffmpeg-go), focused strictly on **forming CLI commands**.

*   **Logic Only**: All execution wrappers and process management have been removed. 
*   **CLI-First**: The library now serves only as a tool for generating `ffmpeg` command-line arguments.
*   **Lightweight**: Stripped of high-level streaming helpers and unnecessary third-party dependencies to keep the codebase minimal.
*   **Manual Execution**: Users are responsible for running the generated command (e.g., via `os/exec`).

# How to get and use
You can get this package via:
```
go get -u github.com/stain4/ffmpeg-cli-go
```

# Examples

```go
in1 := ffmpeg.Input("clip1.mp4")
in2 := ffmpeg.Input("clip2.mp4")
out := ffmpeg.Output([]*ffmpeg.Stream{
	in1.Get("v:0").Crop(0, 0, 1280, 720),
	in2.Get("a:0"),
	ffmpeg.RawArgs("-c:v:0", "libx264", "-preset", "fast"),
	in1.MapChapters(),
	in1.MapMetadata("", "v:0"),
	in2.MapMetadata("s:v:0", "s:v:0"),
	in1.MapMetadata("s:a:0", "s:a:0"),
}, "outfile.mp4", ffmpeg.KwArgs{"c:a": "aac", "b:a": "128K"})
fmt.Printf("%s\n", strings.Join(out.GetArgs(), " "))
```

result:

```bash
-i clip1.mp4 -i clip2.mp4 -filter_complex [0:v:0]crop=1280:720:0:0[s0] -map_chapters 0 -map [s0] -map 1:a:0 -map_metadata 0:v:0 -map_metadata:s:v:0 1:s:v:0 -map_metadata:s:a:0 0:s:a:0 -c:v:0 libx264 -preset fast -b:a 128K -c:a aac outfile.mp4
```
