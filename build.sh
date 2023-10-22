#!/bin/sh

echo "Building UI dist directory"
npm run build --prefix ui/

echo "Compiling Go binary"
go build -o btry -ldflags="-s -w" .