#!/bin/bash

/bin/echo "Creating Agent for Ubuntu"


/bin/mkdir /tmp/ubuntu/
/bin/cp -r lightowl/* /tmp/ubuntu/
/bin/cp ../program/telegraf/telegraf-ubuntu-debian.deb /tmp/ubuntu/
/bin/cp installer.sh /tmp/ubuntu/

cd ../program
env GOOS=linux GOARCH=amd64 go build -o /tmp/ubuntu/etc/lightowl/lightowl ./lightowl.go

cd /tmp/ubuntu/
chmod +x ./installer.sh
/usr/bin/makeself . /tmp/to_upload/lightowl-agent-ubuntu.run "LightOwl Agent Installer" ./installer.sh
