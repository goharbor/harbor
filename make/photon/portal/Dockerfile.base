FROM photon:4.0

RUN tdnf install -y nginx shadow >> /dev/null \
    && tdnf clean all \
    && ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log \
    && groupadd -r -g 10000 nginx && useradd --no-log-init -r -g 10000 -u 10000 nginx \
    && chown -R nginx:nginx /etc/nginx