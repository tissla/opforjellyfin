#!/bin/bash

set -e

echo "Building Linux-binary.."
# Linux 64-bit
GOOS=linux GOARCH=amd64 go build -o opfor-linux

echo "Linux build done!"
echo "Building MacOS-binary.."
# macOS 64-bit (Intel)
GOOS=darwin GOARCH=amd64 go build -o opfor-macos

echo "MacOS build done!"
echo "Building .exe.."
# Windows 64-bit (skapar .exe)
GOOS=windows GOARCH=amd64 go build -o opfor.exe

echo "All done!"
