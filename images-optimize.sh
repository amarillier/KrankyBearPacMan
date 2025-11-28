#!/bin/bash

# Image Optimization Script for KrankyBear PacMan
# Resizes sprite images to reduce binary size
# Original 1024x1024 images are much larger than needed (displayed at 54-64px)

echo "KrankyBear PacMan - Image Optimization Script"
echo "=============================================="
echo ""

# Target size - 256x256 is 4x the display size, plenty of quality headroom
TARGET_SIZE=256

# Check if sips is available (macOS)
if ! command -v sips &> /dev/null; then
    echo "Error: sips not found. This script requires macOS."
    echo "On Linux, install ImageMagick and modify script to use 'convert' instead."
    exit 1
fi

# Directories containing sprite images (not the originals or misc images)
SPRITE_DIRS=(
    "Resources/Images/traditional"
    "Resources/Images/camo"
    "Resources/Images/animal"
    "Resources/Images/football"
    "Resources/Images/hogwarts"
    "Resources/Images/hogwarts2"
    "Resources/Images/collegefootball"
    "Resources/Images/superheroes"
)

echo "Before optimization:"
du -sh Resources/Images
echo ""

# Process each directory
for dir in "${SPRITE_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        echo "Processing $dir..."
        
        # Process only PNG files in the main directory (not original/ subdirectory)
        for img in "$dir"/*.png; do
            if [ -f "$img" ]; then
                # Get current dimensions
                current_size=$(sips -g pixelWidth "$img" 2>/dev/null | tail -1 | awk '{print $2}')
                
                if [ "$current_size" -gt "$TARGET_SIZE" ] 2>/dev/null; then
                    filename=$(basename "$img")
                    echo "  Resizing $filename from ${current_size}px to ${TARGET_SIZE}px"
                    sips -Z $TARGET_SIZE "$img" --out "$img" >/dev/null 2>&1
                fi
            fi
        done
    fi
done

echo ""
echo "After optimization:"
du -sh Resources/Images
echo ""

# Also show individual directories
echo "Directory breakdown:"
for dir in "${SPRITE_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        size=$(du -sh "$dir" 2>/dev/null | cut -f1)
        echo "  $dir: $size"
    fi
done

echo ""
echo "=============================================="
echo "Optimization complete!"
echo ""
echo "Note: Original high-resolution images are preserved in each theme's 'original/' subdirectory."
echo "To rebuild with optimized images, run: go build ."

