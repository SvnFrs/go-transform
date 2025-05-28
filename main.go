package main

import (
	"bytes"
	"encoding/binary"
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

// EncodeICO converts an image to ICO format and writes it to w
func EncodeICO(w *os.File, img image.Image) error {
	// Convert image to PNG first (ICO will contain a PNG)
	pngBuffer := new(bytes.Buffer)
	err := png.Encode(pngBuffer, img)
	if err != nil {
		return err
	}
	pngBytes := pngBuffer.Bytes()
	pngSize := len(pngBytes)

	// Write ICO header
	dir := icondir{
		Reserved: 0,
		Type:     1, // 1 = ICO, 2 = CUR
		Count:    1, // We only embed one image
	}

	err = binary.Write(w, binary.LittleEndian, dir)
	if err != nil {
		return err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Use 0 for 256px dimensions as per ICO format spec
	var widthByte, heightByte byte
	if width == 256 {
		widthByte = 0
	} else {
		widthByte = byte(width)
	}
	if height == 256 {
		heightByte = 0
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
		BitsPerPixel: 32, // PNG with alpha channel
		Size:         uint32(pngSize),
		Offset:       22, // Size of icondir (6) + size of icondirEntry (16) = 22
	}

	err = binary.Write(w, binary.LittleEndian, entry)
	if err != nil {
		return err
	}

	// Write the PNG data
	_, err = w.Write(pngBytes)
	return err
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

func main() {
	// Define command line flags
	inputFile := flag.String("input", "", "Input image file path (required)")
	outputFile := flag.String("output", "", "Output image file path (if not specified, will use input filename with suffix)")
	resizePercent := flag.Int("resize", 0, "Resize percentage (1-99). 0 means no resize")
	compressLevel := flag.Int("compress", 0, "Compression level (1-100, where 1 is max compression, 100 is best quality). 0 means no compression")
	convertToIco := flag.Bool("to-ico", false, "Convert the image to ICO format")

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

	// Determine output filename and path
	var outPath string
	if *outputFile != "" {
		// If output file is specified, use it as-is but ensure it goes to the right folder
		category := determineOutputCategory(*resizePercent, *compressLevel, *convertToIco)
		outputDir := filepath.Join("output", category)

		// Ensure output directory exists
		err = ensureOutputDir(outputDir)
		if err != nil {
			log.Fatalf("Error creating output directory: %v", err)
		}

		filename := filepath.Base(*outputFile)
		if *convertToIco && !strings.HasSuffix(strings.ToLower(filename), ".ico") {
			// Add .ico extension if converting to ICO
			filename += ".ico"
		}
		outPath = filepath.Join(outputDir, filename)
	} else {
		// Generate output filename automatically
		inputBasename := filepath.Base(*inputFile)
		ext := filepath.Ext(inputBasename)
		basename := strings.TrimSuffix(inputBasename, ext)

		suffix := ""
		if *resizePercent > 0 {
			suffix += fmt.Sprintf("_r%d", *resizePercent)
		}
		if *compressLevel > 0 {
			suffix += fmt.Sprintf("_c%d", *compressLevel)
		}

		// Determine output category and directory
		category := determineOutputCategory(*resizePercent, *compressLevel, *convertToIco)
		outputDir := filepath.Join("output", category)

		// Ensure output directory exists
		err = ensureOutputDir(outputDir)
		if err != nil {
			log.Fatalf("Error creating output directory: %v", err)
		}

		// Change extension if converting to ICO
		var filename string
		if *convertToIco {
			filename = basename + suffix + ".ico"
		} else {
			filename = basename + suffix + ext
		}

		outPath = filepath.Join(outputDir, filename)
	}

	// Create output file
	out, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer out.Close()

	// Handle ICO conversion specifically
	if *convertToIco {
		err = EncodeICO(out, img)
		if err != nil {
			log.Fatalf("Error encoding to ICO format: %v", err)
		}
		fmt.Printf("Image converted to ICO format and saved to %s\n", outPath)
		return
	}

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
