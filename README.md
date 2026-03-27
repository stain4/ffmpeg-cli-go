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
