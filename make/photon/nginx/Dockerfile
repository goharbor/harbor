# syntax=docker/dockerfile:1
# check=skip=InvalidDefaultArgInFrom

ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-nginx-base:${harbor_base_image_version}

VOLUME /var/cache/nginx /var/log/nginx /run

USER nginx

STOPSIGNAL SIGQUIT

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl --fail -s http://localhost:8080 || exit 1

CMD ["nginx", "-g", "daemon off;"]
