#!/bin/bash

if [ -z $1 ]
then
    export IMAGE_TAG=$(git branch --show-current)
else
    export IMAGE_TAG=$1
fi

# Make sure the build directory exists
[ ! -d ./build ] && mkdir -p ./build

# Clear the old files from the build directory
[ ! -f ./build/modbus2mqtt-$IMAGE_TAG.tar.gz ] && rm ./build/modbus2mqtt-$IMAGE_TAG.tar.gz

docker build -t modbus2mqtt:$IMAGE_TAG -f Dockerfile .

docker save modbus2mqtt:$IMAGE_TAG | gzip > ./build/modbus2mqtt-$IMAGE_TAG.tar.gz
