package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
)

func main() {
	// Define command line flags
	inputFile := flag.String("input", "", "Input image file path (required)")
	outputFile := flag.String("output", "", "Output image file path (if not specified, will use input filename with suffix)")
	resizePercent := flag.Int("resize", 0, "Resize percentage (1-99). 0 means no resize")
	compressLevel := flag.Int("compress", 0, "Compression level (1-100, where 1 is max compression, 100 is best quality). 0 means no compression")

	flag.Parse()

	// Validate inputs
	if *inputFile == "" {
		log.Fatal("Input file is required. Use -input flag to specify the input image.")
	}

	if *resizePercent < 0 || *resizePercent > 99 {
		log.Fatal("Resize percentage must be between 1 and 99, or 0 for no resizing")
	}

	if *compressLevel < 0 || *compressLevel > 100 {
		log.Fatal("Compression level must be between 1 and 100, or 0 for no compression")
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

	// Process the image - resize if requested
	if *resizePercent > 0 {
		bounds := img.Bounds()
		width := uint(float64(bounds.Dx()) * float64(*resizePercent) / 100.0)
		height := uint(float64(bounds.Dy()) * float64(*resizePercent) / 100.0)

		// Ensure minimum dimensions of 1 pixel
		if width < 1 {
			width = 1
		}
		if height < 1 {
			height = 1
		}

		img = resize.Resize(width, height, img, resize.Lanczos3)
		fmt.Printf("Image resized to %d%% (%dx%d pixels)\n", *resizePercent, width, height)
	}

	// Determine output filename if not provided
	outPath := *outputFile
	if outPath == "" {
		ext := filepath.Ext(*inputFile)
		basename := strings.TrimSuffix(*inputFile, ext)

		suffix := ""
		if *resizePercent > 0 {
			suffix += fmt.Sprintf("_r%d", *resizePercent)
		}
		if *compressLevel > 0 {
			suffix += fmt.Sprintf("_c%d", *compressLevel)
		}

		outPath = basename + suffix + ext
	}

	// Create output file
	out, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer out.Close()

	// Save the processed image with compression if applicable
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		var opts jpeg.Options
		if *compressLevel > 0 {
			opts.Quality = *compressLevel
		} else {
			opts.Quality = 95 // default quality
		}
		err = jpeg.Encode(out, img, &opts)
		if *compressLevel > 0 {
			fmt.Printf("Image compressed with quality level %d\n", *compressLevel)
		}

	case "png":
		encoder := png.Encoder{}
		if *compressLevel > 0 {
			// For PNG, higher compression level means more compression (opposite of JPEG)
			// Convert our 1-100 scale (where 1 is max compression) to PNG's 0-9 scale (where 9 is max compression)
			level := png.CompressionLevel(9 - int(float64(*compressLevel)/100.0*9.0))
			encoder.CompressionLevel = level
			fmt.Printf("Image compressed with PNG compression level %v\n", level)
		}
		err = encoder.Encode(out, img)

	default:
		// For other formats, just encode without compression options
		err = png.Encode(out, img)
	}

	if err != nil {
		log.Fatalf("Error encoding output image: %v", err)
	}

	fmt.Printf("Processed image saved to %s\n", outPath)
}
