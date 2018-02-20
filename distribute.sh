#!/usr/bin/env bash

name=plexStats
export CGO_ENABLED=1

platforms=(
    linux-amd64 windows-amd64
    linux-386 windows-386
    linux-armv5 linux-armv6 linux-armv7
    #linux-armv8
)

checkCommand() {
    which "$1" >/dev/null 2>&1
    if [ "$?" != "0" ]; then
        echo Please make sure the following command is available: "$1" >&2
        exit "$?"
    fi
}

checkCommand go
checkCommand tar
checkCommand zip

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd "$DIR"
commit="$(git rev-parse --short HEAD 2>/dev/null)"
date="$(date --iso-8601=seconds)"

if [ "$commit" == "" ]; then
    commit="unknown"
fi

if [ "$1" == "" ]; then
    version=$(git tag | tail -1)
    echo You didn\'t provide a version string as the first parameter, setting version to the current git tag \"${version}\".
else
    version="$1"
fi

rm -Rf ./bin/
mkdir ./bin/ 2>/dev/null

for plat in "${platforms[@]}"; do
    export GOOS="${plat%-*}"
    export GOARCH="${plat#*-}"

    if [ "$GOOS" != "windows" ]; then
        tmpFile="/tmp/${name}/bin/${name}"
    else
        tmpFile="/tmp/${name}/bin/${name}.exe"
    fi

    if [ "$CGO_ENABLED" == 1 ]; then
        if [ "$plat" = "windows-amd64" ]; then
            export CC=x86_64-w64-mingw32-gcc
        elif [ "$plat" = "windows-386" ]; then
            export CC=i686-w64-mingw32-gcc
        elif [ "${GOARCH%v*}" = "arm" ]; then
            export GOARM="${GOARCH#*v}"
            if [ "GOARM" = 8 ]; then
                unset GOARM=""
                export CC=armv8l-linux-gnueabihf-gcc
                export GOARCH="arm64"
            else
                export CC=arm-linux-gnueabihf-gcc
                export GOARCH="arm"
            fi
        else
            export CC=gcc
        fi
    fi

    echo Building "$plat" with "$CC" ...
    go build -ldflags '-extld='"$CC"' -X main.VERSION='"$version"' -X main.BUILD_COMMIT='"$commit"' -X main.BUILD_DATE='"$date" \
    -o "$tmpFile" "$DIR"/*.go

    if [ "$?" != 0 ]; then
        echo Build failed! >&2
        exit "$?"
    fi

    if [ "$GOOS" != "windows" ]; then
        tarPath="$DIR"/bin/${name}-"$plat".tar.gz
        echo Build succeeded, creating "$tarPath" ...
        tar -czf "$tarPath" -C "${tmpFile%/*}" ${name}
    else
        zipPath="$DIR"/bin/${name}-"$plat".zip
        echo Build succeeded, creating "$zipPath" ...
        zip -j "$zipPath" "$tmpFile"
    fi

    if [ "$?" != 0 ]; then
        echo Failed to pack the binary! >&2
        exit "$?"
    fi
    echo Done!

    rm "$tmpFile"

    echo
done