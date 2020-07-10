#!/usr/bin/env bash
set -e

PACKAGE="voiceip-b2bua"

function die() {
    echo $1;
    exit 1
}

function replacePlaceHolders() {
    file="$1"
    $SED_CMD -i -e "s/_PACKAGE_/$PACKAGE/g" $file
}


if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "Build Starting...."
else
    die "Build only supported on linux os"
fi


SED_CMD="sed"
# if [ "$(uname)" == "Darwin" ]; then
#     SED_CMD="gsed"
# fi


BUILD_ROOT=$(mktemp -d)
VERSION=$(date +%s)

cp -r distribution/debian/* $BUILD_ROOT/

mkdir -p $BUILD_ROOT/usr/local/bin/
mkdir -p bin

##build
GOOS=linux go build -gcflags "all=-N -l" -o bin/voiceip-b2bua.linux 
cp ./bin/voiceip-b2bua.linux  $BUILD_ROOT/usr/local/bin/$PACKAGE


#replacing constants
replacePlaceHolders "${BUILD_ROOT}/DEBIAN/prerm"
replacePlaceHolders "${BUILD_ROOT}/DEBIAN/postrm"
replacePlaceHolders "${BUILD_ROOT}/DEBIAN/postinst"
replacePlaceHolders "${BUILD_ROOT}/DEBIAN/control"

$SED_CMD -i "s/_VERSION_/$VERSION/g" $BUILD_ROOT/DEBIAN/control

echo "Package Version: 1.$VERSION"

rm -f $PACKAGE.deb
dpkg-deb --build $BUILD_ROOT $PACKAGE.deb

rm -rf $BUILD_ROOT

