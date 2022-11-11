#!/bin/bash

# Copyright (c) 2022 Contributors to the Eclipse Foundation
#
# See the NOTICE file(s) distributed with this work for additional
# information regarding copyright ownership.
#
# This program and the accompanying materials are made available under the
# terms of the Apache License 2.0 which is available at
# https://www.apache.org/licenses/LICENSE-2.0
#
# SPDX-License-Identifier: Apache-2.0

platforms=("linux/amd64/linux-x86_64" "linux/arm/linux-arm" "linux/arm64/linux-arm64" "windows/amd64/windows-x86_64" "darwin/amd64/macos-x86_64")
ldflags_value="-s -w"
[ "$CLOUD_CONNECTOR_VERSION" ] && ldflags_value="$ldflags_value -X 'main.version=$CLOUD_CONNECTOR_VERSION'"
ldflags="-ldflags=$ldflags_value"
otherflags="-trimpath -mod=readonly"
artifactId="cloud_connector"
mkdir -p target
for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}

    if [ "$GOOS" = "windows" ]; then
        binary_name=cloudconnector.exe
    else
        binary_name=cloudconnector
    fi

    echo "build_module: target platform [$platform]"
    target_folder="target/natives/${platform_split[2]}"
    mkdir -p "$target_folder"
    env GOOS=$GOOS GOARCH=$GOARCH go build -o $target_folder/$binary_name "$ldflags" $otherflags

    if [ $? -ne 0 ]; then
        echo "build_module: [$1][Error]: Build for platform [$platform] failed! Will exit (1)!"
        exit 1
    fi

    if [ "$GOOS" = "linux" ]; then
        chmod 0755 $target_folder/$binary_name
    fi

	cp -R ../LICENSE $target_folder
	cp ../NOTICE.md $target_folder
	cp ../resources/iothub.crt $target_folder
    cp -R ../resources/protobuf $target_folder
    cp ../resources/message-mapper-config.json $target_folder
    if [ "$GOOS" = "windows" ]; then
	    cp  ../resources/cloudconnector_start.bat $target_folder
	    cp  ../resources/cloudconnector_stop.bat $target_folder
    else
	    cp ../resources/cloudconnector_start.sh $target_folder
	    cp ../resources/cloudconnector_stop.sh $target_folder
    fi

	cd $target_folder && tar --exclude='.' -czf ../../$artifactId-${platform_split[2]}.tar.gz * && cd ../../..
done

cd target && tar -czf $artifactId-package.tar.gz $artifactId-linux*.tar.gz $artifactId-windows*.tar.gz $artifactId-macos*.tar.gz && cd ..
