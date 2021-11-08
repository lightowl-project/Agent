#/bin/bash

set -e

if [[ $EUID -ne 0 ]]; then
   /bin/echo "This script must be run as root" 
   exit 1
fi

if [ $# -eq 0 ]; then
    echo "Usage: $0 <lightowl_ip_address> <agent_token>"
    exit 1
fi

/usr/bin/yum install -y curl unzip hostname

SERVER_ADDR=$1
API_KEY=$2
HOSTNAME="`/usr/bin/hostname`"

if [ -x "$(command -v telegraf)" ]; then
   echo "Telegraf already installed"
else
    /usr/bin/rpm -i telegraf.rpm
fi

# /usr/bin/wget -qO- https://repos.influxdata.com/influxdb.key | /usr/bin/sudo apt-key add -
# /bin/echo "Installing Telegraf"
# source /etc/lsb-release
# /bin/echo "deb https://repos.influxdata.com/${DISTRIB_ID,,} ${DISTRIB_CODENAME} stable" | /usr/bin/sudo tee /etc/apt/sources.list.d/influxdb.list
# apt update && apt install -y telegraf

if id "lightowl" &>/dev/null; then
   deluser lightowl
fi

/usr/sbin/adduser --system lightowl --disabled-password --no-create-home
/usr/sbin/adduser lightowl telegraf

/usr/bin/cp -r ./etc /
/usr/bin/mkdir -p /var/log/lightowl
/usr/bin/mkdir -p /etc/ssl/lightowl/
/usr/bin/chown -R lightowl:root /var/log/lightowl

OS="CentOS"

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
/usr/bin/unzip ./lightowl.zip
cd /tmp/lightowl

/usr/bin/mv ./ca.pem /etc/ssl/lightowl/
/usr/bin/mv ./telegraf.conf /etc/telegraf/telegraf.conf
/usr/bin/mv ./lightowl.conf /etc/telegraf/telegraf.d/lightowl.conf

/usr/bin/rm /tmp/lightowl.zip
/usr/bin/rm -rf /tmp/lightowl

(/usr/bin/crontab -l -u lightowl; echo "* * * * * /etc/lightowl/lightowl-linux-amd64") | awk '!x[$0]++' | sudo crontab -u lightowl -
