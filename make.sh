#!/bin/bash

ROOT=${PWD}
FILES="main.go handlers.go"

go build -C src -o blog ${FILES}
if [[ $? -ne 0 ]]
then
	echo "build failed"
	exit
fi
mv src/blog bin