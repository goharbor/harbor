FROM vmware/photon:1.0

RUN tdnf distro-sync -y \
    && tdnf install -y nginx \
    && ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log \
    && mkdir -p /var/run \
    && tdnf clean all

EXPOSE 80
STOPSIGNAL SIGQUIT

CMD ["nginx", "-g", "daemon off;"]
