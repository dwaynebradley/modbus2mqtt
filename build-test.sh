#!/bin/bash

# Make sure the build directory exists
[ ! -d ./build ] && mkdir -p ./build

# Clear the old files from the build directory
[ -f ./build/modbus2mqtt-test ] && rm ./build/modbus2mqtt-test
# [ -f ./build/modbus2mqtt-test.exe ] && rm ./build/modbus2mqtt-test.exe

# Linux x64 build
echo -n "Building Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -o ./build/modbus2mqtt-test .
echo "Done!"

# Windows x64 build
# echo -n "Building Windows AMD64..."
# GOOS=windows GOARCH=amd64 go build -o ./build/modbus2mqtt-test.exe .
# echo "Done!"
