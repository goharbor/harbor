#!/bin/bash
# Requires: openssl

source ./harbor.cfg

config_path="./config"
ui_path="./config/ui"
db_path="./config/db"
tpl_path="./templates"

mkdir -p $tpl_path
mkdir -p $ui_path $db_path

declare -a arr=("$ui_path/env" "$ui_path/app.conf" "$config_path/registry/config.yml" "$db_path/env")

for i in "${arr[@]}"; do
    if [ -e $i ]; then
        echo "Clearing the configuration file: "$i
        rm $i
    fi
done

source $tpl_path/ui/app.conf > $ui_path/app.conf
echo "Generated configuration file: "$ui_path/app.conf

source $tpl_path/ui/env > $ui_path/env
echo "Generated configuration file: "$ui_path/env

source $tpl_path/db/env > $config_path/db/env
echo "Generated configuration file: "$config_path/db/env

source $tpl_path/registry/config.yml > $config_path/registry/config.yml
echo "Generated configuration file: "$config_path/registry/config.yml

is_fail=0

if [ $customize_token == "on" ];then

    if [ -e $ui_path/private_key.pem ]; then
        echo "clearing the origin private_key.pem in "$ui_pth
        rm $ui_path/private_key.pem
    fi
    openssl genrsa -out $ui_path/private_key.pem 4096
    if [ -e $ui_path/private_key.pem ]; then
        echo "private_key.gem has been generated in "$ui_path
    else echo "generate private_key.gem fail."
        is_fail=1
    fi

    if [ -e $config_path/registry/root.crt ]; then
        echo "clearing the origin root.crt in "$config_path"/registry"
        rm $config_path/registry/root.crt
    fi

    openssl req -new -x509 -key $ui_path/private_key.pem -out $config_path/registry/root.crt -days 3650 \
        -subj "/C=$crt_countryname/ST=$crt_state/L=$crt_name/O=$crt_organizationname/OU=$crt_organizationalunitname"
    if [ -e $config_path/registry/root.crt ]; then
        echo "root.crt has been generated in "$config_path"/registry"
    else echo "generate root.crt fail."
        is_fail=1
    fi
elif [ $customize_token != "off" ]; then
    echo "wrong args found in customize_token: "$customize_token
    is_fail=1
fi

if [ $is_fail -eq 0 ];then
    echo "The configuration files are ready, please use docker-compose to start the service."
else
    echo "some problem occurs."
fi
