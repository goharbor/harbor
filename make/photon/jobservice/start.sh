#!/bin/sh

sudo -E -u \#10000 "/harbor/harbor_jobservice" "-c" "/etc/jobservice/config.yml"

