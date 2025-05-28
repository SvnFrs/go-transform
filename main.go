package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
)

// ICO file format structures
type icondir struct {
	Reserved uint16
	Type     uint16
	Count    uint16
}

type icondirEntry struct {
	Width        byte
	Height       byte
	PaletteCount byte
	Reserved     byte
	ColorPlanes  uint16
	BitsPerPixel uint16
	Size         uint32
	Offset       uint32
}

// convertToRGBA ensures the image is in RGBA format
func convertToRGBA(src image.Image) *image.RGBA {
	if rgba, ok := src.(*image.RGBA); ok {
		return rgba
	}

	bounds := src.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, src, bounds.Min, draw.Src)
	return rgba
}

// resizeForICO resizes image for ICO format if needed
func resizeForICO(img image.Image, maxSize int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// If image is already within limits, return as-is
	if width <= maxSize && height <= maxSize {
		return img
	}

	// Calculate new dimensions maintaining aspect ratio
	var newWidth, newHeight uint
	if width > height {
		newWidth = uint(maxSize)
		newHeight = uint(float64(height) * float64(maxSize) / float64(width))
	} else {
		newHeight = uint(maxSize)
		newWidth = uint(float64(width) * float64(maxSize) / float64(height))
	}

	// Ensure minimum dimensions
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	resized := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
	fmt.Printf("Image resized for ICO format: %dx%d -> %dx%d\n", width, height, newWidth, newHeight)
	return resized
}

// EncodeICO converts an image to ICO format and writes it to w
func EncodeICO(w *os.File, img image.Image, autoResize bool) error {
	// Auto-resize if requested and image is too large
	if autoResize {
		img = resizeForICO(img, 256)
	}

	// Ensure the image is in RGBA format
	rgbaImg := convertToRGBA(img)

	// Create PNG encoder with best compression for smaller ICO files
	pngBuffer := new(bytes.Buffer)
	encoder := &png.Encoder{
		CompressionLevel: png.BestCompression,
	}

	err := encoder.Encode(pngBuffer, rgbaImg)
	if err != nil {
		return fmt.Errorf("failed to encode PNG for ICO: %w", err)
	}

	pngBytes := pngBuffer.Bytes()
	pngSize := len(pngBytes)

	// Write ICO header
	dir := icondir{
		Reserved: 0,
		Type:     1, // 1 = ICO, 2 = CUR
		Count:    1, // We only embed one image
	}

	if err := binary.Write(w, binary.LittleEndian, dir); err != nil {
		return fmt.Errorf("failed to write ICO header: %w", err)
	}

	bounds := rgbaImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Handle dimensions larger than 255 (modern ICO format support)
	var widthByte, heightByte byte
	if width >= 256 {
		widthByte = 0 // 0 means 256 in ICO format
	} else {
		widthByte = byte(width)
	}
	if height >= 256 {
		heightByte = 0 // 0 means 256 in ICO format
	} else {
		heightByte = byte(height)
	}

	// Write ICO directory entry
	entry := icondirEntry{
		Width:        widthByte,
		Height:       heightByte,
		PaletteCount: 0,
		Reserved:     0,
		ColorPlanes:  1,
		BitsPerPixel: 32, // 32-bit RGBA
		Size:         uint32(pngSize),
		Offset:       22, // Size of icondir (6) + size of icondirEntry (16) = 22
	}

	if err := binary.Write(w, binary.LittleEndian, entry); err != nil {
		return fmt.Errorf("failed to write ICO directory entry: %w", err)
	}

	// Write the PNG data
	if _, err := w.Write(pngBytes); err != nil {
		return fmt.Errorf("failed to write PNG data to ICO: %w", err)
	}

	return nil
}

// determineOutputCategory determines which output folder to use based on operations
func determineOutputCategory(resizePercent int, compressLevel int, convertToIco bool) string {
	if convertToIco {
		return "transform"
	}
	if resizePercent > 0 {
		return "resize"
	}
	if compressLevel > 0 {
		return "compress"
	}
	return "processed" // fallback for any other processing
}

// ensureOutputDir creates the output directory if it doesn't exist
func ensureOutputDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// validateFlags validates command line arguments
func validateFlags(inputFile *string, resizePercent *int, compressLevel *int) error {
	if *inputFile == "" {
		return fmt.Errorf("input file is required. Use -input flag to specify the input image")
	}

	if *resizePercent < 0 || *resizePercent > 99 {
		return fmt.Errorf("resize percentage must be between 1 and 99, or 0 for no resizing")
	}

	if *compressLevel < 0 || *compressLevel > 100 {
		return fmt.Errorf("compression level must be between 1 and 100, or 0 for no compression")
	}

	// Check if input file exists
	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", *inputFile)
	}

	return nil
}

// resizeImage resizes the image if needed
func resizeImage(img image.Image, resizePercent int) (image.Image, error) {
	if resizePercent <= 0 {
		return img, nil
	}

	bounds := img.Bounds()
	width := uint(float64(bounds.Dx()) * float64(resizePercent) / 100.0)
	height := uint(float64(bounds.Dy()) * float64(resizePercent) / 100.0)

	// Ensure minimum dimensions of 1 pixel
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	resized := resize.Resize(width, height, img, resize.Lanczos3)
	fmt.Printf("Image resized to %d%% (%dx%d pixels)\n", resizePercent, width, height)
	return resized, nil
}

// generateOutputPath generates the output file path
func generateOutputPath(inputFile, outputFile string, resizePercent, compressLevel int, convertToIco bool) (string, error) {
	var outPath string

	if outputFile != "" {
		// If output file is specified, use it as-is but ensure it goes to the right folder
		category := determineOutputCategory(resizePercent, compressLevel, convertToIco)
		outputDir := filepath.Join("output", category)

		// Ensure output directory exists
		if err := ensureOutputDir(outputDir); err != nil {
			return "", fmt.Errorf("error creating output directory: %w", err)
		}

		filename := filepath.Base(outputFile)
		if convertToIco && !strings.HasSuffix(strings.ToLower(filename), ".ico") {
			// Add .ico extension if converting to ICO
			filename += ".ico"
		}
		outPath = filepath.Join(outputDir, filename)
	} else {
		// Generate output filename automatically
		inputBasename := filepath.Base(inputFile)
		ext := filepath.Ext(inputBasename)
		basename := strings.TrimSuffix(inputBasename, ext)

		suffix := ""
		if resizePercent > 0 {
			suffix += fmt.Sprintf("_r%d", resizePercent)
		}
		if compressLevel > 0 {
			suffix += fmt.Sprintf("_c%d", compressLevel)
		}

		// Determine output category and directory
		category := determineOutputCategory(resizePercent, compressLevel, convertToIco)
		outputDir := filepath.Join("output", category)

		// Ensure output directory exists
		if err := ensureOutputDir(outputDir); err != nil {
			return "", fmt.Errorf("error creating output directory: %w", err)
		}

		// Change extension if converting to ICO
		var filename string
		if convertToIco {
			filename = basename + suffix + ".ico"
		} else {
			filename = basename + suffix + ext
		}

		outPath = filepath.Join(outputDir, filename)
	}

	return outPath, nil
}

// encodeImage handles encoding the image in the appropriate format
func encodeImage(out *os.File, img image.Image, format string, compressLevel int) error {
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		var opts jpeg.Options
		if compressLevel > 0 {
			opts.Quality = compressLevel
		} else {
			opts.Quality = 95 // default quality
		}

		if err := jpeg.Encode(out, img, &opts); err != nil {
			return fmt.Errorf("failed to encode JPEG: %w", err)
		}

		if compressLevel > 0 {
			fmt.Printf("Image compressed with quality level %d\n", compressLevel)
		}

	case "png":
		encoder := png.Encoder{}
		if compressLevel > 0 {
			// For PNG, higher compression level means more compression (opposite of JPEG)
			// Convert our 1-100 scale (where 1 is max compression) to PNG's 0-9 scale (where 9 is max compression)
			level := png.CompressionLevel(9 - int(float64(compressLevel)/100.0*9.0))
			encoder.CompressionLevel = level
			fmt.Printf("Image compressed with PNG compression level %v\n", level)
		}

		if err := encoder.Encode(out, img); err != nil {
			return fmt.Errorf("failed to encode PNG: %w", err)
		}

	default:
		// For other formats, just encode as PNG
		if err := png.Encode(out, img); err != nil {
			return fmt.Errorf("failed to encode as PNG: %w", err)
		}
	}

	return nil
}

func main() {
	// Define command line flags
	inputFile := flag.String("input", "", "Input image file path (required)")
	outputFile := flag.String("output", "", "Output image file path (if not specified, will use input filename with suffix)")
	resizePercent := flag.Int("resize", 0, "Resize percentage (1-99). 0 means no resize")
	compressLevel := flag.Int("compress", 0, "Compression level (1-100, where 1 is max compression, 100 is best quality). 0 means no compression")
	convertToIco := flag.Bool("to-ico", false, "Convert the image to ICO format")
	autoResizeICO := flag.Bool("auto-resize-ico", true, "Automatically resize images larger than 256x256 when converting to ICO")

	flag.Parse()

	// Validate inputs
	if err := validateFlags(inputFile, resizePercent, compressLevel); err != nil {
		log.Fatal(err)
	}

	// Open the input file
	file, err := os.Open(*inputFile)
	if err != nil {
		log.Fatalf("Error opening input file: %v", err)
	}
	defer file.Close()

	// Decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		log.Fatalf("Error decoding image: %v", err)
	}

	fmt.Printf("Loaded %s image: %dx%d\n", format, img.Bounds().Dx(), img.Bounds().Dy())

	// Process the image - resize if requested
	img, err = resizeImage(img, *resizePercent)
	if err != nil {
		log.Fatalf("Error resizing image: %v", err)
	}

	// Generate output path
	outPath, err := generateOutputPath(*inputFile, *outputFile, *resizePercent, *compressLevel, *convertToIco)
	if err != nil {
		log.Fatalf("Error generating output path: %v", err)
	}

	// Create output file
	out, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			log.Printf("Warning: Error closing output file: %v", closeErr)
		}
	}()

	// Handle ICO conversion specifically
	if *convertToIco {
		// Show warning for large images if auto-resize is disabled
		bounds := img.Bounds()
		if (bounds.Dx() > 256 || bounds.Dy() > 256) && !*autoResizeICO {
			log.Printf("Warning: Large image dimensions (%dx%d) may not display properly in all ICO viewers. Consider using -auto-resize-ico=true", bounds.Dx(), bounds.Dy())
		}

		if err := EncodeICO(out, img, *autoResizeICO); err != nil {
			log.Fatalf("Error encoding to ICO format: %v", err)
		}
		fmt.Printf("Image converted to ICO format (RGBA) and saved to %s\n", outPath)
		return
	}

	// Save the processed image with compression if applicable
	if err := encodeImage(out, img, format, *compressLevel); err != nil {
		log.Fatalf("Error encoding output image: %v", err)
	}

	fmt.Printf("Processed image saved to %s\n", outPath)
}
