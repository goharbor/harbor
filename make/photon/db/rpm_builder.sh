#!/bin/bash

set -e

name='postgresql'
version='9.6.21'

function checkdep {
	if ! wget --version &> /dev/null
	then
		echo "Need to install wget first and run this script again."
		exit 1
	fi

	if ! [ -x "$(command -v bzip2)" ]
	then
		echo "Need to install bzip2 first and run this script again."
		exit 1
	fi
}

checkdep

cur=$PWD
workDir=`mktemp -d ${TMPDIR-/tmp}/$name.XXXXXX`
mkdir -p $workDir && cd $workDir

# step 1: get source code of pg 9.6, and rename the code directory from postgres to postgres96
wget http://ftp.postgresql.org/pub/source/v$version/$name-$version.tar.bz2
bzip2 -d ./$name-$version.tar.bz2 && tar -xvf ./$name-$version.tar
mkdir -p ${name}96-$version && cp -r ./$name-$version/* ./${name}96-$version/ && rm -rf ./$name-$version
tar -cvjSf ${name}96-$version.tar.bz2 ${name}96-$version

# step 2: get spec builder script, and replace version to 4, then to build the pg96 rpm packages
wget https://raw.githubusercontent.com/vmware/photon/4.0/tools/scripts/build_spec.sh
sed "s|VERSION=3|VERSION=4|g" -i build_spec.sh
chmod 655 ./build_spec.sh && cp $cur/postgres.spec .
./build_spec.sh ./postgres.spec
cp ./stage/RPMS/x86_64/${name}96-libs-$version-1.ph4.x86_64.rpm $cur
cp ./stage/RPMS/x86_64/${name}96-$version-1.ph4.x86_64.rpm $cur

# clean
cd $cur && rm -rf $workDir
