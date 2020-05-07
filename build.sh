#!/bin/bash

rm -rf build/*

GOOS=windows GOARCH=amd64 go build -o build/windows_amd64/golivereload.exe
cd build/windows_amd64
zip `pwd | xargs basename`.zip *
mv *.zip ..
cd -

GOOS=linux GOARCH=amd64 go build -o build/linux_amd64/golivereload
cd build/linux_amd64
tar czf `pwd | xargs basename`.tar.gz *
mv *.tar.gz ..
cd -

GOOS=darwin go build -o build/macos/golivereload
cd build/macos
tar czf `pwd | xargs basename`.tar.gz *
mv *.tar.gz ..
cd -
