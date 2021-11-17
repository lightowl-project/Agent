#!/bin/bash

/bin/echo "Creating Agent for Debian"

/bin/mkdir /tmp/debian/
/bin/cp -r lightowl/* /tmp/debian/
/bin/cp ../program/telegraf/telegraf-ubuntu-debian.deb /tmp/debian/telegraf.deb
/bin/cp installer.sh /tmp/debian/

cd ../program
ls -la
env GOOS=linux GOARCH=amd64 go build -o /tmp/debian//etc/lightowl/lightowl ./lightowl.go


cd /tmp/debian/
chmod +x ./installer.sh
/usr/bin/makeself . /tmp/to_upload/lightowl-agent-debian.run "LightOwl Agent Installer" ./installer.sh
