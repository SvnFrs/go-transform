# Go Image Processor

A command-line tool for resizing, compressing, and converting images. Supports common image formats and can convert images to ICO format for Windows icons with proper RGBA support.

## Features

- **Resize images** by percentage (1-99%)
- **Compress images** with adjustable quality levels (1-100)
- **Convert to ICO format** for Windows icons with RGBA support
- **Auto-resize for ICO** - automatically resize large images for optimal ICO compatibility
- **Auto-generate output filenames** with descriptive suffixes
- **Organized output folders** - automatically categorizes processed images
- **Support for multiple formats**: JPEG, PNG, and ICO
- **Input validation** - checks file existence and parameter ranges
- **Proper error handling** with detailed error messages

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
- `-to-ico`: Convert the image to ICO format with RGBA support
- `-auto-resize-ico`: Automatically resize images larger than 256x256 when converting to ICO (default: true)

### Examples

**Basic resize:**
```bash
./img-processor -input image.jpg -resize 50
# Output: output/resize/image_r50.jpg (50% of original size)
```

**Compress image:**
```bash
./img-processor -input photo.jpg -compress 75
# Output: output/compress/photo_c75.jpg (75% quality)
```

**Resize and compress:**
```bash
./img-processor -input large.png -resize 25 -compress 80
# Output: output/resize/large_r25_c80.png
```

**Convert to ICO (with auto-resize):**
```bash
./img-processor -input logo.png -to-ico
# Output: output/transform/logo.ico (auto-resized to ≤256x256 if needed)
```

**Convert to ICO (preserve large dimensions):**
```bash
./img-processor -input logo.png -to-ico -auto-resize-ico=false
# Output: output/transform/logo.ico (keeps original dimensions)
```

**Convert large favicon:**
```bash
./img-processor -input favicon.png -to-ico
# Loaded png image: 512x512
# Image resized for ICO format: 512x512 -> 256x256
# Image converted to ICO format (RGBA) and saved to output/transform/favicon.ico
```

**Custom output filename:**
```bash
./img-processor -input image.jpg -output thumbnail.jpg -resize 30
# Output: output/resize/thumbnail.jpg
```

## Output Organization

The tool automatically organizes output files into folders based on the operation:

- `output/resize/` - Images that were resized
- `output/compress/` - Images that were compressed
- `output/transform/` - Images converted to ICO format
- `output/processed/` - Other processed images

## Compression Quality

- **JPEG**: 1 = lowest quality/smallest file, 100 = highest quality/largest file
- **PNG**: Uses PNG's built-in compression levels (automatically converted from 1-100 scale)

## ICO Format Features

When converting to ICO format:
- **RGBA Support**: Ensures proper alpha channel handling for transparency
- **Auto-resize**: Large images (>256x256) are automatically resized for better compatibility
- **Quality preservation**: Uses optimal PNG compression within ICO container
- **Modern compatibility**: Supports both traditional and modern ICO viewers
- **Aspect ratio preservation**: Smart resizing maintains original proportions

### ICO Best Practices

- **Recommended sizes**: 16x16, 32x32, 48x48, 128x128, 256x256
- **Auto-resize**: Enabled by default for images larger than 256x256
- **Transparency**: Fully supported with proper RGBA encoding
- **Quality**: High-quality Lanczos3 resampling for resizing

## Error Handling

The tool provides comprehensive error checking:
- Input file existence validation
- Parameter range validation
- Detailed error messages with context
- Graceful handling of unsupported formats
- Warning messages for suboptimal operations

## Dependencies

- [github.com/nfnt/resize](https://github.com/nfnt/resize) - High-quality image resizing with Lanczos3 algorithm

## Supported Formats

- **Input**: JPEG, PNG, GIF, BMP, TIFF, and other formats supported by Go's image package
- **Output**: JPEG, PNG, ICO

## File Naming Convention

When output filename is not specified, the tool automatically generates names with suffixes:
- `_r{percentage}` for resize operations
- `_c{level}` for compression operations
- Combined: `filename_r50_c75.jpg`

## Technical Details

- **RGBA Conversion**: All images are converted to RGBA format when creating ICO files
- **PNG Embedding**: ICO files contain high-quality PNG data
- **Memory Efficient**: Processes images without loading multiple copies into memory
- **Cross-platform**: Works on Windows, macOS, and Linux

## Troubleshooting

**Large ICO files**: If your ICO file is too large, the auto-resize feature will automatically reduce dimensions to 256x256 or smaller.

**Format compatibility**: The tool automatically detects input format and preserves it for output (except when converting to ICO).

**Permission errors**: Ensure you have write permissions in the directory where the tool creates the `output` folder.
