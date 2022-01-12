#!/bin/bash

version=$1
/bin/echo "Creating Agent for Windows"

/bin/mkdir /tmp/windows/
/bin/cp -r lightowl/* /tmp/windows/
/bin/cp ../program/telegraf/telegraf-1.21.1_windows_amd64.zip /tmp/windows/telegraf.zip
/bin/cp installer.ps1 /tmp/windows/

cd ../program
env GOOS=windows GOARCH=amd64 go build -o /tmp/windows/etc/lightowl/lightowl.exe ./lightowl-windows.go

cd /tmp/
/usr/bin/zip -r /tmp/to_upload/lightowl-agent-windows-amd64-$version.zip ./windows/
# msi-packager ./to_upload/windows/ /tmp/to_upload/lightowl-agent-windows.msi -e installer.ps1 