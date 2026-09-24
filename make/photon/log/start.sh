#!/bin/bash
set -e
chown -R 10000:10000 /var/log/docker
crond
rm -f /var/run/rsyslogd.pid
exec rsyslogd -n

