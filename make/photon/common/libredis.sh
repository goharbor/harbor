#!/bin/sh

set -e

_redis_cred() {
    if [ -z "${REDIS_USERNAME}" ] && [ -z "${REDIS_PASSWORD}" ]; then
      return 0
    fi

    echo -n "${REDIS_USERNAME}:${REDIS_PASSWORD}@"
}

_sentinel_master_set() {
    case "$REDIS_SCHEME" in
        *+sentinel)
            echo -n "/${REDIS_MASTER_SET}"
            ;;
    esac
}

_redis_url() {
    # $1 is db index
    echo -n "$REDIS_SCHEME://$(_redis_cred)${REDIS_ADDR}$(_sentinel_master_set)/$1"
}

_precheck() {
    if [ -z "${REDIS_SCHEME}" ] && [ -z "${REDIS_ADDR}" ]; then
        echo "Using default ${1:-_REDIS_URL}_* variables for redis"
        return 1
    fi
}

_validate_redis() {
    case "$REDIS_SCHEME" in
        redis|rediss)
            # valid
            ;;
        redis+sentinel|rediss+sentinel)
            if [ -z "${REDIS_MASTER_SET}" ]; then
                echo "Error: REDIS_MASTER_SET not set, but sentinel is enabled" >&2
            fi
            ;;
        *)
            echo "Error: invalid REDIS_SCHEME: $REDIS_SCHEME" >&2
            exit 1
            ;;
    esac

    if [ -z "${REDIS_ADDR}" ]; then
        echo "Error: REDIS_SCHEME is present but REDIS_ADDR is empty" >&2
        exit 1
    fi
}

_configure_redis_core() {
    if [ -z "${_REDIS_URL_CORE}" ] && [ -z "${_REDIS_URL_HARBOR}" ] && [ -z "${REDIS_HARBOR_DB_INDEX}" ]; then
        echo "ERROR: _REDIS_URL_CORE, _REDIS_URL_HARBOR and REDIS_HARBOR_DB_INDEX are not set, configure at least one" >&2
        exit 1
    fi

    # NOTE: _REDIS_URL_HARBOR is not ever set by this script
    # It can still be set by user and will overwrite configuration done by this script
    if [ -z "${_REDIS_URL_CORE}" ]; then
        echo "Using REDIS_* variables for harbor/cache"
        export _REDIS_URL_CORE="$(_redis_url "${REDIS_HARBOR_DB_INDEX}")"
    else
        echo "Using _REDIS_URL_CORE for harbor/cache"
    fi

    if [ ! -z "${_REDIS_URL_HARBOR}" ]; then
        echo "Using _REDIS_URL_HARBOR for harbor, will override _REDIS_URL_CORE"
        # NOTE: _REDIS_URL_CORE might still be used for cache if cache is enabled and _REDIS_URL_CACHE_LAYER is not configured
    fi
}

_configure_redis_reg() {
    if [ -z "${_REDIS_URL_REG}" ]; then
        echo "Using REDIS_* variables for registry controller"

        if [ -z "${REDIS_REG_DB_INDEX}" ]; then
            echo "ERROR: _REDIS_URL_REG and REDIS_REG_DB_INDEX are not set, configure at least one" >&2
            exit 1
        fi

        export _REDIS_URL_REG="$(_redis_url "${REDIS_REG_DB_INDEX}")"
    else
        echo "Using _REDIS_URL_REG for registry controller"
    fi
}

_configure_redis_cache() {
    # Here do not fail if REDIS_CACHE_DB_INDEX is empty since cache is optional
    if [ -z "${_REDIS_URL_CACHE_LAYER}" ]; then
        if [ ! -z "${REDIS_CACHE_DB_INDEX}" ]; then
            echo "Using REDIS_* variables for cache layer"
            export _REDIS_URL_CACHE_LAYER="$(_redis_url "${REDIS_CACHE_DB_INDEX}")"
        else
            echo "Not configuring redis url for cache layer, REDIS_CACHE_DB_INDEX is not set"
        fi
    else
        echo "Using _REDIS_URL_CACHE_LAYER for cache layer"
    fi
}

configure_redis_core() {
    echo "Configuring Redis for Harbor core..."

    _precheck || return 0

    _validate_redis

    _configure_redis_core

    _configure_redis_reg

    _configure_redis_cache

}

configure_redis_jobservice() {
    echo "Configuring Redis for Harbor jobservice..."

    _precheck || return 0

    _validate_redis

    _configure_redis_core

    _configure_redis_cache
}

configure_redis_trivy() {
    echo "Configuring Redis for Trivy..."

    _precheck "SCANNER" || return 0

    _validate_redis

    if [ -z "$SCANNER_REDIS_URL" ] && [ -z "$SCANNER_STORE_REDIS_URL" ] && [ -z "$SCANNER_JOB_QUEUE_REDIS_URL" ]; then
        echo "Using REDIS_* variables for Trivy"

        if [ -z "${REDIS_TRIVY_DB_INDEX}" ]; then
            echo "ERROR: No url for redis set in SCANNER_* variables, REDIS_TRIVY_DB_INDEX not set, configure either one" >&2
            exit 1
        fi

        export SCANNER_REDIS_URL="$(_redis_url "${REDIS_TRIVY_DB_INDEX}")"
        export SCANNER_STORE_REDIS_URL="$(_redis_url "${REDIS_TRIVY_DB_INDEX}")"
        export SCANNER_JOB_QUEUE_REDIS_URL="$(_redis_url "${REDIS_TRIVY_DB_INDEX}")"
    else
        echo "Using SCANNER_* variables for Trivy redis"
    fi
}
