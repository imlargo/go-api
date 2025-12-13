# Repurposer CLI

A command-line tool for repurposing videos locally without using cloud storage or database.

## Overview

This CLI tool allows you to process videos using the Butter repurposer engine directly from the command line. It applies various transformations to videos (rotation, color correction, overlays, etc.) and saves the output locally to your filesystem.

## Usage

### Build the CLI

```bash
cd backend
go build -o ./tmp/repurposer-cli ./cmd/repurposer-cli
```

### Run the CLI

Basic usage:
```bash
./tmp/repurposer-cli <input_video_path> [output_directory]
```

Examples:
```bash
# Process video and save to default ./output directory
./tmp/repurposer-cli /path/to/video.mp4

# Process video and save to specific directory
./tmp/repurposer-cli /path/to/video.mp4 /path/to/output
```

## Configuration

The repurposer options are configured directly in the code. To modify the transformation settings, edit the `renderOptions` in `main.go`:

```go
renderOptions := repurposer.RenderOptions{
    InputPath:   inputPath,
    Template:    nil, // Use default template
    UseMirror:   true,
    UseOverlays: true,
    TextOverlay: "Repurposed by Butter CLI",
    IsMain:      false,
}
```

### Available Options

- **UseMirror**: Apply mirror effect to the video
- **UseOverlays**: Add overlay effects
- **TextOverlay**: Custom text to display as overlay
- **IsMain**: Flag to indicate if this is the main account processing

## Output

The CLI generates two files:
1. **Repurposed video**: The processed video with all transformations applied (MP4 format)
2. **Thumbnail**: A thumbnail image extracted from the repurposed video

Both files are saved to the specified output directory (default: `./output`).

## Requirements

- FFmpeg must be installed and available in your system PATH
- Go 1.21 or later

## Notes

- This CLI tool does NOT save files to cloud storage
- This CLI tool does NOT interact with the database
- All processing is done locally
- The repurposer engine uses the same transformation pipeline as the web service
