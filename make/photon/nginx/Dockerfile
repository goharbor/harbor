FROM photon:2.0

RUN tdnf install -y nginx >> /dev/null\
    && ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log \
    && tdnf clean all

EXPOSE 80
VOLUME /var/cache/nginx /var/log/nginx /run
STOPSIGNAL SIGQUIT

HEALTHCHECK CMD curl --fail -s http://127.0.0.1 || exit 1

CMD ["nginx", "-g", "daemon off;"]
