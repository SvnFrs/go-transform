# Go Image Processor

A command-line tool for resizing, compressing, and converting images. Supports common image formats and can convert images to ICO format for Windows icons.

## Features

- **Resize images** by percentage (1-99%)
- **Compress images** with adjustable quality levels (1-100)
- **Convert to ICO format** for Windows icons
- **Auto-generate output filenames** with descriptive suffixes
- **Support for multiple formats**: JPEG, PNG, and ICO

## Installation

Make sure you have Go installed, then:

```bash
go mod tidy
go build -o img-processor
```

## Usage

```bash
./img-processor [flags]
```

### Flags

- `-input` (required): Input image file path
- `-output` (optional): Output image file path. If not specified, generates filename with suffix
- `-resize`: Resize percentage (1-99). 0 means no resize
- `-compress`: Compression level (1-100, where 1 is max compression, 100 is best quality). 0 means no compression
- `-to-ico`: Convert the image to ICO format

### Examples

**Basic resize:**
```bash
./img-processor -input image.jpg -resize 50
# Output: image_r50.jpg (50% of original size)
```

**Compress image:**
```bash
./img-processor -input photo.jpg -compress 75
# Output: photo_c75.jpg (75% quality)
```

**Resize and compress:**
```bash
./img-processor -input large.png -resize 25 -compress 80
# Output: large_r25_c80.png
```

**Convert to ICO:**
```bash
./img-processor -input logo.png -to-ico
# Output: logo.ico
```

**Custom output filename:**
```bash
./img-processor -input image.jpg -output thumbnail.jpg -resize 30
```

## Compression Quality

- **JPEG**: 1 = lowest quality/smallest file, 100 = highest quality/largest file
- **PNG**: Uses PNG's built-in compression levels (automatically converted from 1-100 scale)

## ICO Format

When converting to ICO format:
- The image is embedded as PNG data within the ICO container
- Supports transparency and high quality
- Suitable for Windows application icons

## Dependencies

- [github.com/nfnt/resize](https://github.com/nfnt/resize) - High-quality image resizing

## Supported Formats

- **Input**: JPEG, PNG, and other formats supported by Go's image package
- **Output**: JPEG, PNG, ICO

## File Naming Convention

When output filename is not specified, the tool automatically generates names with suffixes:
- `_r{percentage}` for resize operations
- `_c{level}` for compression operations
- Combined: `filename_r50_c75.jpg`
