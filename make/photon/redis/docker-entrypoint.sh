#!/bin/bash
set -e

REDIS_PARAMETER=''

if [ -n "${REDIS_PASSWORD}" ]; then
    REDIS_PARAMETER="${REDIS_PARAMETER} --requirepass "${REDIS_PASSWORD}"
fi

exec redis-server /etc/redis.conf ${REDIS_PARAMETER}