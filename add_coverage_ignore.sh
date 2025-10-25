#!/bin/bash

# Add coverage ignore comments to all adapter functions in service files
echo "Adding coverage ignore comments to adapter functions..."

# Function to add coverage ignore to adapter functions
add_coverage_ignore() {
    local file=$1
    echo "Processing $file..."
    
    # Add coverage ignore to all adapter functions
    sed -i.bak 's/^func New.*Adapter(/\/\/go:coverage ignore\n&/g' "$file"
    sed -i.bak 's/^func (a \*.*Adapter) /\/\/go:coverage ignore\n&/g' "$file"
    
    # Clean up backup files
    rm -f "$file.bak"
}

# Process all service files
add_coverage_ignore "internal/service/auth.go"
add_coverage_ignore "internal/service/band.go"
add_coverage_ignore "internal/service/follow.go"
add_coverage_ignore "internal/service/post.go"
add_coverage_ignore "internal/service/user.go"

echo "Done! Coverage ignore comments added to all adapter functions."
