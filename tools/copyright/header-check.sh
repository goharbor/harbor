#!/bin/bash
# Copyright Project Harbor Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Simple script that will check files of type .go, .sh, .bash or Makefile
# for the copyright header.
#
# This will be called by the CI system (with no args) to perform checking and
# fail the job if headers are not correctly set. It can also be called with the
# 'fix' argument to automatically add headers to the missing files.
#
# Check if headers are fine:
#   $ ./scripts/header-check.sh
# Check and fix headers:
#   All changes must be committed for fix to work
#   $ ./scripts/header-check.sh fix

set -e -o pipefail

# These header variables MUST match the first two lines of the
# copyright file in the scripts directory.
#
# These will be evaluated as a regex against the target file
HEADER_OLD[1]="^\/\*$"
HEADER_OLD[2]="^\s*Copyright \(c\) [0-9]{4} The Harbor Authors$"

HEADER_NEW[1]="^\/\/ Copyright \(c\) [0-9]{4} The Harbor Authors$"
HEADER_NEW[2]="^\/\/$"

# Initialize vars
ERR=false
FAIL=false
OLD_MATCHED=false

all-files() {
    git ls-files |\
        # Check .go files, Makefile, sh files, bash files, and robot files
        # grep -e "\.go$" -e "\.ts$" -e "\.css$" -e "Makefile$" -e "\.sh$" -e "\.bash$" -e "\.robot$" |\
        grep -e "\.sh$" |\
            # Ignore vendor/
        grep -v vendor |\
            # Ignore api test
        grep -v apitests
        # / |\
        #     # Ignore files marked as deleted in git
        # grep -v "$(git status -s | grep -e "^\sD" | cut -d' ' -f3)"
}

for file in $(all-files); do
  ext=${file##*.}
  if [[ $ext == "go" ]]; then
    echo -n "Header check: $file... "
    # get the file extension / type
    OLD_MATCHED=false

    # match the old style header
    for count in $(seq 1 ${#HEADER_OLD[@]}); do
      text=${HEADER_OLD[$count]}
      if [[ "$increment" = true ]]; then
        line=$((count + 1))
      else
        line=$count
      fi
      # do we have a header match?
      if [[ $(sed ${line}q\;d ${file}) =~ ${text} ]]; then
        OLD_MATCHED=true
      fi
    done

    # add header if there isn't.
    if [ $OLD_MATCHED == true ]; then
      # is there is a fix argument and are all changes committed
      if [ "$(uname -s)" = "Darwin" ]; then
        permissions=$(stat -f "%OLp" ${file})
      else
        permissions=$(stat --format '%a' ${file})
      fi

      # make it writable
      chmod 755 ${file}

      # remove the old header.
      tail -n +15 ${file} > ${file}.new

      mv ${file}.new ${file}

      # make permissions the same
      chmod $permissions ${file}
      echo "$(tput -T xterm setaf 3)REMOVING OLD HEADER $(tput -T xterm sgr0)"
    else
      echo "$(tput -T xterm setaf 3)NO OLD HEADER $(tput -T xterm sgr0)"
    fi
  fi
done

sleep 2

for file in $(all-files); do
  echo -n "Header check: $file... "
  # get the file extension / type
  ext=${file##*.}

  # increment line count in certain cases
  increment=false

  # should we be incrementing the line count
  if [[ $ext == "sh" ]]; then
    increment=true
  fi

  # match the new style header
  for count in $(seq 1 ${#HEADER_NEW[@]}); do
    text=${HEADER_NEW[$count]}

    if [[ "$increment" = true ]]; then
      line=$((count + 1))
    else
      line=$count
    fi
    # do we have a header match?
    if [[ ! $(sed ${line}q\;d ${file}) =~ ${text} ]]; then
      ERR=true
    fi
  done

  # add header if there isn't.
  if [ $ERR == true ]; then
    # is there is a fix argument and are all changes committed
    if [[ $# -gt 0 && $1 =~ [[:upper:]fix] ]]; then
      if [ "$(uname -s)" = "Darwin" ]; then
        permissions=$(stat -f "%OLp" ${file})
      else
        permissions=$(stat --format '%a' ${file})
      fi

      # make it writable
      chmod 755 ${file}

      # based on file type fix the copyright
      case "$ext" in
        go)
          cat ./tools/copyright/copyright ${file} > ${file}.new
          ;;
        ts)
          cat ./tools/copyright/copyright ${file} > ${file}.new
          ;;
        sh)
          head -1 ${file} > ${file}.new
          cat ./tools/copyright/copyright | sed 's/\/\//\#/1' >> ${file}.new
          tail -n +2 ${file} >> ${file}.new
          ;;
        *)
          cat ./tools/copyright/copyright | sed 's/\/\//\#/1' > ${file}.new
          cat ${file} >> ${file}.new
          ;;
      esac

      mv ${file}.new ${file}

      # make permissions the same
      chmod $permissions ${file}

      echo "$(tput -T xterm setaf 3)FIXING$(tput -T xterm sgr0)"
      ERR=false
    else
      echo "$(tput -T xterm setaf 1)FAIL$(tput -T xterm sgr0)"
      ERR=false
      FAIL=true
    fi
  else
    echo "$(tput -T xterm setaf 2)OK$(tput -T xterm sgr0)"
  fi
done

# If we failed one check, return 1
[ $FAIL == true ] && exit 1 || exit 0
