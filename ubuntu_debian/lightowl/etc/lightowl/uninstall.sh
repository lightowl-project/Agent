#!/bin/bash

if [[ $EUID -ne 0 ]]; then
   /bin/echo "This script must be run as root" 
   exit 1
fi

/bin/systemctl stop telegraf
/usr/bin/apt remove -y telegraf

/usr/sbin/deluser lightowl

/bin/rm -rf /etc/lightowl
/bin/rm -rf /etc/telegraf
/bin/rm -rf /etc/ssl/lightowl
/bin/rm -rf /var/log/lightowl
/bin/rm -f /etc/sudoers.d/lightowl