#!/bin/bash
set -e

VERSION="${1:-1.0.0}"
ARCH="${2:-amd64}"
PACKAGE_NAME="tux"
PACKAGE_DIR="${PACKAGE_NAME}_${VERSION}_${ARCH}"
OUTPUT="${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"

echo "Building .deb package: ${OUTPUT}"

# Clean up any previous build
rm -rf "${PACKAGE_DIR}"

# Create directory structure
mkdir -p "${PACKAGE_DIR}/usr/bin"
mkdir -p "${PACKAGE_DIR}/usr/lib/systemd/user"
mkdir -p "${PACKAGE_DIR}/DEBIAN"

# Build binaries
echo "Building binaries..."
cd "$(dirname "$0")/../.."
go build -o "${PACKAGE_DIR}/usr/bin/tux" ./cmd/tux
go build -o "${PACKAGE_DIR}/usr/bin/tuxd" ./cmd/tuxd

# Install systemd service
cp packaging/tux.service "${PACKAGE_DIR}/usr/lib/systemd/user/"

# Create control file
sed "s/\${VERSION}/${VERSION}/g" packaging/deb/control > "${PACKAGE_DIR}/DEBIAN/control"

# Calculate installed size
INSTALLED_SIZE=$(du -sk "${PACKAGE_DIR}" | cut -f1)
sed -i "s/Installed-Size: 10/Installed-Size: ${INSTALLED_SIZE}/" "${PACKAGE_DIR}/DEBIAN/control"

# Set permissions
chmod 0755 "${PACKAGE_DIR}/usr/bin/tux"
chmod 0755 "${PACKAGE_DIR}/usr/bin/tuxd"
chmod 0644 "${PACKAGE_DIR}/usr/lib/systemd/user/tux.service"

# Build the package
dpkg-deb --build "${PACKAGE_DIR}" "${OUTPUT}"

# Clean up
rm -rf "${PACKAGE_DIR}"

echo "Built ${OUTPUT} successfully!"
