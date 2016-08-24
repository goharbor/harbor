#!/bin/bash
set -e
echo "This shell will minify the Javascript in Harbor project."
echo "Usage: #jsminify [src] [dest] [basedir]"

#prepare workspace
rm -rf $2 /tmp/harbor.app.temp.js

if [ -z $3 ] 
then
    BASEPATH=/go/bin
else
    BASEPATH=$3
fi

#concat the js files from js include file
echo "Concat js files..."

cat $1   | while read LINE || [[ -n $LINE ]]
do
    if [ -n "$LINE" ] 
    then
        TEMP="$BASEPATH""$LINE"
        cat `echo "$TEMP" | sed 's/<script src=\"//g' | sed 's/\"><\/script>//g'` >> /tmp/harbor.app.temp.js
        printf "\n" >> /tmp/harbor.app.temp.js
    fi
done

# If you want run this script on Mac OS X,
# I suggest you install gnu-sed (whth --with-default-names option).
# $ brew install gnu-sed --with-default-names
# Reference:
# http://stackoverflow.com/a/27834828/3167471

#remove space
echo "Remove space.."
sed 's/ \+/ /g' -i /tmp/harbor.app.temp.js

#remove '//' and '/*'
echo "Remove '//'and '/*'  annotation..."
sed '/^\/\//'d -i /tmp/harbor.app.temp.js
sed '/\/\*/{/\*\//d;:a;N;/\*\//d;ba};s,//.*,,' -i /tmp/harbor.app.temp.js 

cat > $2 << EOF
/*
    Copyright (c) 2016 VMware, Inc. All Rights Reserved.
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
        
        http://www.apache.org/licenses/LICENSE-2.0
        
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/
EOF

#remove '\n'
echo "Remove CR  ..."
cat /tmp/harbor.app.temp.js | tr -d '\n' >> $2

#clear workspace
rm -rf /tmp/harbor.app.temp.js

echo "Done."
exit  0

 
