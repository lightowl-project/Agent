#!/bin/bash

/bin/echo "Creating Agent for Ubuntu"

/bin/mkdir ./tmp/
/bin/cp -r lightowl/* ./tmp/
/bin/cp telegraf.deb ./tmp/
env GOOS=linux GOARCH=amd64 go build -o ./tmp/etc/lightowl/lightowl ./lightowl.go

/bin/cp installer.sh ./tmp/
cd ./tmp/

ls -la ./etc/lightowl

/usr/bin/makeself . ../../lightowl-agent-ubuntu.run "LightOwl Agent Installer" ./installer.sh

ls -la ../../
