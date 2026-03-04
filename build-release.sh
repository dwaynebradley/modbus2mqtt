#!/bin/bash

# Make sure the build directory exists
[ ! -d ./build ] && mkdir -p ./build

# Clear the old files from the build directory
[ -f ./build/modbus2mqtt ] && rm ./build/modbus2mqtt

# Linux x64 build
echo -n "Building..."
CGO_ENABLED=0 go build -a -ldflags="-s -w" -o ./build/modbus2mqtt .
echo "Done!"

