#!/bin/bash
set -e

VERSION="${1:-1.0.0}"
ARCH="${2:-x86_64}"
PACKAGE_NAME="tux"
SOURCE_TAR="${PACKAGE_NAME}-${VERSION}.tar.gz"
SPEC_FILE="${PACKAGE_NAME}.spec"
RPMBUILD_DIR="${HOME}/rpmbuild"
OUTPUT="${PACKAGE_NAME}-${VERSION}-1.${ARCH}.rpm"

echo "Building .rpm package: ${OUTPUT}"

# Clean up any previous build
rm -f "${SOURCE_TAR}"

# Create source tarball
echo "Creating source tarball..."
TAR_FILES="cmd/ internal/ go.mod packaging/"
[ -f "go.sum" ] && TAR_FILES="${TAR_FILES} go.sum"

tar --transform "s,^,${PACKAGE_NAME}-${VERSION}/," \
    --exclude='*.git*' \
    --exclude='*_test.go' \
    -czf "${SOURCE_TAR}" \
    ${TAR_FILES}

# Create rpmbuild directory structure
mkdir -p "${RPMBUILD_DIR}"/{SOURCES,SPECS,BUILD,RPMS,SRPMS}

# Move source tarball
mv "${SOURCE_TAR}" "${RPMBUILD_DIR}/SOURCES/"

# Process spec file with version and date
DATE=$(date "+%a %b %d %Y")
sed "s/\${VERSION}/${VERSION}/g" packaging/rpm/tux.spec | \
    sed "s/\${DATE}/${DATE}/g" > "${RPMBUILD_DIR}/SPECS/${SPEC_FILE}"

# Build the package
echo "Building RPM package..."
rpmbuild -ba "${RPMBUILD_DIR}/SPECS/${SPEC_FILE}" \
    --target "${ARCH}-linux" \
    --define "_sourcedir ${RPMBUILD_DIR}/SOURCES" \
    --define "_specdir ${RPMBUILD_DIR}/SPECS" \
    --define "_builddir ${RPMBUILD_DIR}/BUILD" \
    --define "_rpmdir ${RPMBUILD_DIR}/RPMS"

# Copy the built package to current directory
cp "${RPMBUILD_DIR}/RPMS/${ARCH}/${OUTPUT}" .

echo "Built ${OUTPUT} successfully!"
