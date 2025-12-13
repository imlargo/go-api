package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/nicolailuther/butter/pkg/ffmpeg"
	"github.com/nicolailuther/butter/pkg/files"
	"github.com/nicolailuther/butter/pkg/repurposer"
	"github.com/nicolailuther/butter/pkg/transform"
)

// copyFile copies a file from src to dst, handling cross-filesystem moves
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Sync to ensure all data is written to disk
	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	// Remove the source file after successful copy
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("failed to remove source file: %w", err)
	}

	return nil
}

func main() {
	// Usage: go run cmd/repurposer-cli/main.go <input_video_path> [output_directory]
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <input_video_path> [output_directory]\n", os.Args[0])
		fmt.Println("\nExample:")
		fmt.Printf("  %s /path/to/video.mp4\n", os.Args[0])
		fmt.Printf("  %s /path/to/video.mp4 /path/to/output\n", os.Args[0])
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputDir := "./output"
	if len(os.Args) >= 3 {
		outputDir = os.Args[2]
	}

	// Validate input file exists
	if err := files.CheckFile(inputPath); err != nil {
		log.Fatalf("Input file error: %v", err)
	}

	// Ensure output directory exists
	if err := files.EnsureDirectoryExists(outputDir, true); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	fmt.Println("🎬 Butter CLI Repurposer")
	fmt.Println("========================")
	fmt.Printf("Input:  %s\n", inputPath)
	fmt.Printf("Output: %s\n\n", outputDir)

	// Initialize FFmpeg client
	ffmpegClient, err := ffmpeg.New(&ffmpeg.Options{})
	if err != nil {
		log.Fatalf("Failed to initialize FFmpeg client: %v", err)
	}

	// Initialize repurposer engine
	engine, err := repurposer.New(ffmpegClient, transform.DefaultParameters(), repurposer.DefaultOptions())
	if err != nil {
		log.Fatalf("Failed to create repurposer engine: %v", err)
	}

	// Manually construct the DTO with repurposer options
	// These settings can be modified directly in the code for different effects
	renderOptions := repurposer.RenderOptions{
		InputPath:   inputPath,
		Template:    nil, // Use default template
		UseMirror:   false,
		UseOverlays: true,
		TextOverlay: "Guess who’s living rent-free in your head again?",
		IsMain:      false,
	}

	fmt.Println("⚙️  Render Options:")
	fmt.Printf("  - Use Mirror:   %v\n", renderOptions.UseMirror)
	fmt.Printf("  - Use Overlays: %v\n", renderOptions.UseOverlays)
	fmt.Printf("  - Text Overlay: %q\n", renderOptions.TextOverlay)
	fmt.Printf("  - Is Main:      %v\n\n", renderOptions.IsMain)

	// Process the video
	fmt.Println("🔄 Processing video...")
	result, err := engine.RepurposeVideo(context.Background(), renderOptions)
	if err != nil {
		log.Fatalf("Failed to repurpose video: %v", err)
	}

	// Move the output file to the desired output directory
	outputFilename := filepath.Base(result.OutputPath)
	finalOutputPath := filepath.Join(outputDir, outputFilename)

	if err := copyFile(result.OutputPath, finalOutputPath); err != nil {
		log.Fatalf("Failed to move output file: %v", err)
	}

	fmt.Println("\n✅ Video repurposed successfully!")
	fmt.Printf("📁 Output: %s\n", finalOutputPath)
	fmt.Printf("⏱️  Processing time: %.2f seconds\n\n", result.Duration.Seconds())

	// Generate thumbnail
	fmt.Println("📸 Generating thumbnail...")
	thumbnailResult, err := engine.RenderThumbnail(repurposer.RenderOptions{
		InputPath: finalOutputPath,
		Template:  nil,
	})
	if err != nil {
		log.Fatalf("Failed to render thumbnail: %v", err)
	}

	// Move the thumbnail to the output directory
	thumbnailFilename := filepath.Base(thumbnailResult.OutputPath)
	finalThumbnailPath := filepath.Join(outputDir, thumbnailFilename)

	if err := copyFile(thumbnailResult.OutputPath, finalThumbnailPath); err != nil {
		log.Fatalf("Failed to move thumbnail file: %v", err)
	}

	fmt.Println("✅ Thumbnail generated successfully!")
	fmt.Printf("📁 Thumbnail: %s\n", finalThumbnailPath)
	fmt.Printf("⏱️  Processing time: %.2f seconds\n\n", thumbnailResult.Duration.Seconds())

	fmt.Println("🎉 All done! Your repurposed content is ready.")
}
