#!/bin/bash
set -x
set -e
profile_path=$(pwd)/profile.cov
echo "profile.cov path: $profile_path"

echo "mode: set" >>"$profile_path"

deps=""
cd $(dirname $(find . -name go.mod))
set +x
# listDeps lists packages referenced by package in $1, 
# excluding golang standard library
function listDeps(){
	pkg=$1
	deps=$pkg
	ds=$(echo $(go list -f '{{.Imports}}' $pkg) | sed 's/[][]//g')
	for d in $ds
	do
		if echo $d | grep -q "github.com/goharbor/harbor" && echo $d
		then
			deps="$deps,$d"
		fi
	done
}

packages=$(go list ./... | grep -v -E 'tests|testing')
echo testing packages: $packages

for package in $packages
do
	listDeps $package

#    echo "DEBUG: testing package $package"
	echo go test -race -v -cover -coverprofile=profile.tmp -coverpkg "$deps" $package
	go test -race -v -cover -coverprofile=profile.tmp -coverpkg "$deps" $package
	if [ -f profile.tmp ]	
	then
		cat profile.tmp | tail -n +2 >> "$profile_path"
		rm profile.tmp
	fi	
done
