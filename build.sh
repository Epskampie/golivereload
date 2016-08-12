#!/bin/bash
GOOS=windows GOARCH=amd64 go build -o build/windows_amd64/golivereload.exe
GOOS=windows GOARCH=amd64 go build -o build/linux_amd64/golivereload
GOOS=darwin GOARCH=amd64 go build -o build/osx_amd64/golivereload

