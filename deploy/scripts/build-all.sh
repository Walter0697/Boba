#!/bin/bash

# Build script for creating multi-platform binaries
# This script builds BOBA for all supported platforms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get version from git tag or use dev version
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")

echo -e "${GREEN}Building BOBA v${VERSION}${NC}"
echo "Build time: ${BUILD_TIME}"
echo "Git commit: ${GIT_COMMIT}"
echo ""

# Create dist directory
DIST_DIR="dist"
rm -rf "${DIST_DIR}"
mkdir -p "${DIST_DIR}"

# Build flags
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

# Platforms to build
declare -A PLATFORMS=(
    ["linux-amd64"]="linux amd64"
    ["linux-arm64"]="linux arm64"
    ["darwin-amd64"]="darwin amd64"
    ["darwin-arm64"]="darwin arm64"
    ["windows-amd64"]="windows amd64"
)

# Build for each platform
for platform in "${!PLATFORMS[@]}"; do
    IFS=' ' read -r GOOS GOARCH <<< "${PLATFORMS[$platform]}"
    
    output_name="boba-${platform}"
    if [ "$GOOS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo -e "${YELLOW}Building for ${GOOS}/${GOARCH}...${NC}"
    
    GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build \
        -ldflags="${LDFLAGS}" \
        -o "${DIST_DIR}/${output_name}" \
        . || {
            echo -e "${RED}Failed to build for ${GOOS}/${GOARCH}${NC}"
            exit 1
        }
    
    echo -e "${GREEN}✓ Built ${output_name}${NC}"
done

echo ""
echo -e "${GREEN}Creating archives...${NC}"

# Create archives
cd "${DIST_DIR}"

for file in boba-*; do
    if [[ "$file" == *.exe ]]; then
        # Windows: create zip
        zip_name="${file%.exe}.zip"
        zip -q "$zip_name" "$file"
        echo -e "${GREEN}✓ Created ${zip_name}${NC}"
    else
        # Unix: create tar.gz
        tar_name="${file}.tar.gz"
        tar -czf "$tar_name" "$file"
        echo -e "${GREEN}✓ Created ${tar_name}${NC}"
    fi
done

# Generate checksums
echo ""
echo -e "${GREEN}Generating checksums...${NC}"
sha256sum *.tar.gz *.zip > checksums.txt
echo -e "${GREEN}✓ Created checksums.txt${NC}"

cd ..

echo ""
echo -e "${GREEN}Build complete!${NC}"
echo "Artifacts are in the ${DIST_DIR}/ directory:"
ls -lh "${DIST_DIR}"

echo ""
echo -e "${YELLOW}To test a binary:${NC}"
echo "  ./${DIST_DIR}/boba-linux-amd64 --version"
echo ""
echo -e "${YELLOW}To create a release:${NC}"
echo "  git tag -a v${VERSION} -m 'Release v${VERSION}'"
echo "  git push origin v${VERSION}"
