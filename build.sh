#!/bin/bash


get_telegraf_package() {
    if [[ $1 == "windows" ]]; then
        package_name="telegraf-1.20.0_windows_amd64.zip"
    elif [[ $1 == "linux" ]]; then
        package_name="telegraf_1.20.0-1_amd64.deb"
    elif [[ $1 == "centos" ]]; then
        package_name="telegraf-1.20.0-1.x86_64.rpm"
    fi
}

package=$1
if [[ -z "$package" ]]; then
  echo "usage: $0 <package-name>"
  exit 1
fi
# package_split=(${package//\// })
# package_name=${package_split[-1]}

platforms=("windows/amd64" "linux/amd64")
#, "centos/amd64")

mkdir ./agents/
mkdir ./archives/

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    mkdir "./agents/lightowl-$GOOS/"

    get_telegraf_package $GOOS

    echo "Creating Agent for OS: $GOOS ARCH: $GOARCH Telegraf: $package_name"

    output_name=lightowl-agent'-'$GOOS'-'$GOARCH
    telegraf_name=telegraf
    if [[ $GOOS == "windows" ]]; then
        output_name+='.exe'
        telegraf_name+='.exe'
    elif [[ $GOOS == "linux" ]]; then
        telegraf_name+=".deb"
        cp ./installer_linux.sh ./agents/lightowl-$GOOS/
    elif [[ $GOOS == "centos" ]]; then
        telegraf_name+=".rpm"
        cp ./installer_centos.sh ./agents/lightowl-$GOOS/
    fi

    download_url="https://dl.influxdata.com/telegraf/releases/$package_name"
    wget -q -O "./agents/lightowl-$GOOS/$telegraf_name"  $download_url
    cp -r ./lightowl/* ./agents/lightowl-$GOOS/

    env GOOS=$GOOS GOARCH=$GOARCH go build -o ./agents/lightowl-$GOOS/etc/lightowl/$output_name $package

    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi

    cd ./agents/
    if [[ $GOOS != "windows" ]]; then
        cd ./lightowl-$GOOS/
        makeself . lightowl-$GOOS-$GOARCH.run "LightOwl Agent Installer" ./installer_$GOOS.sh
        mv ./lightowl-$GOOS-$GOARCH.run ../../archives/
        cd ..
    else
        zip -qr ../archives/lightowl-$GOOS'-'$GOARCH'.zip' ./lightowl-$GOOS/
    fi
    cd ..
done

ls -la ./archives/
