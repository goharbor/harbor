#!/bin/bash
set -e
echo "Log rotate starting..."

#The logs n days before will be compressed.
n=14
path=/var/log/docker

list=""
n_days_before=$(($(date +%s) - 3600*24*$n))
for dir in $(ls $path | grep -v "tar.gz");
do
	if [ $(date --date=$dir +%s) -lt $n_days_before ]
	then
		echo "$dir will be compressed"
		list="$list $dir"
	fi
done

if [ -n "$list" ]
then
	cd $path
	tar --remove-files -zcvf $(date -d @$n_days_before +%F)-.tar.gz $list
fi

echo "Log rotate finished."
