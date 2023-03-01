#!/bin/bash

# Make sure the build directory exists
[ ! -d ./build ] && mkdir -p ./build

# Clear the old files from the build directory
[ -f ./build/modbus2mqtt ] && rm ./build/modbus2mqtt
# [ -f ./build/modbus2mqtt.exe ] && rm ./build/modbus2mqtt.exe

# Linux x64 build
echo -n "Building Linux AMD64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o ./build/modbus2mqtt .
echo "Done!"

# Windows x64 build
# echo -n "Building Windows AMD64..."
# CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -installsuffix cgo -o ./build/modbus2mqtt.exe .
# echo "Done!"
