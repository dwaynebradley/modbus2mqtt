#!/bin/bash

# Make sure the build directory exists
[ ! -d ./build ] && mkdir -p ./build

# Clear the old files from the build directory
[ -f ./build/modbus2mqtt-test ] && rm ./build/modbus2mqtt-test

# Linux x64 build
echo -n "Building..."
go build -o ./build/modbus2mqtt-test .
echo "Done!"

