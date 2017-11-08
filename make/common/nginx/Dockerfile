FROM vmware/photon:1.0

RUN tdnf distro-sync -y || echo \
    && tdnf install -y nginx \
    && ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log \
    && mkdir -p /var/run \
    && tdnf clean all

ENV PORT 80
EXPOSE 80
STOPSIGNAL SIGQUIT

HEALTHCHECK CMD curl localhost:$PORT || exit 1
CMD ["nginx", "-g", "daemon off;"]
