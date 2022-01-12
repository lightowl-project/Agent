#!/bin/bash

if [[ $EUID -ne 0 ]]; then
   /bin/echo "This script must be run as root" 
   exit 1
fi

if [ $# -eq 0 ]; then
    echo "Usage: $0 <lightowl_ip_address> <agent_token>"
    exit 1
fi

/usr/bin/apt update
/usr/bin/apt install -y curl unzip

SERVER_ADDR=$1
API_KEY=$2
HOSTNAME="`/bin/hostname`"

if [ -x "$(command -v telegraf)" ]; then
   echo "Telegraf already installed"
else
    /usr/bin/dpkg -i telegraf.deb
fi

if id "lightowl" &>/dev/null; then
   /usr/sbin/deluser lightowl
fi

/usr/sbin/adduser --system lightowl --disabled-password --no-create-home
/usr/sbin/groupadd lightowl
/usr/sbin/adduser lightowl lightowl
/usr/sbin/adduser telegraf lightowl

/bin/cp -r ./etc /
/bin/mkdir -p /var/log/lightowl
/bin/touch /var/log/lightowl/lightowl.log
/bin/mkdir -p /etc/ssl/lightowl/

OS="Linux"

DATA='{
   "os": "'"$OS"'",
   "hostname": "'"$HOSTNAME"'",
   "tags": [],
   "plugins": {}
}';

/usr/bin/curl  -k --location --request POST 'https://'$SERVER_ADDR'/api/v1/agents/join' \
--header "api_key: $API_KEY" \
--header 'Content-Type: application/json' \
--data-raw "$DATA" \
--output /tmp/lightowl.zip

cd /tmp
/usr/bin/unzip /tmp/lightowl.zip
cd ./lightowl

/bin/mv ./.env /etc/lightowl/
/bin/mv ./ca.pem /etc/ssl/lightowl/
/bin/mv ./telegraf.conf /etc/telegraf/telegraf.conf
/bin/mv ./lightowl.conf /etc/telegraf/telegraf.d/lightowl.conf

cd /

/bin/rm /tmp/lightowl.zip
/bin/rm -rf /tmp/lightowl

/bin/chown -R lightowl:root /var/log/lightowl
/bin/chown -R lightowl:root /etc/lightowl
/bin/chown -R lightowl:telegraf /etc/ssl/lightowl
/bin/chmod -R 550 /etc/ssl/lightowl
/bin/chown -R lightowl:root /etc/telegraf

/usr/bin/crontab -u lightowl -l
/usr/bin/crontab -u lightowl -l; echo "* * * * * /etc/lightowl/lightowl" | awk '!x[$0]++' | /usr/bin/crontab -u lightowl -
/usr/bin/crontab -u lightowl -l; echo "0 8 * * * /etc/lightowl/lightowl packages" | awk '!x[$0]++' | /usr/bin/crontab -u lightowl -
/usr/sbin/service telegraf restart

# Updating installed software
sudo -u lightowl /etc/lightowl/lightowl packages
