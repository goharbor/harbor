FROM node:10.15.0 as nodeportal

RUN mkdir -p /portal_src && mkdir -p /build_dir

COPY make/photon/portal/entrypoint.sh /
COPY src/portal /portal_src
COPY ./docs/swagger.yaml   /portal_src

WORKDIR /portal_src

RUN npm install && \
    chmod u+x /entrypoint.sh
RUN /entrypoint.sh


FROM photon:2.0

COPY --from=nodeportal /build_dir/dist /usr/share/nginx/html
COPY --from=nodeportal /build_dir/swagger.yaml /usr/share/nginx/html
COPY --from=nodeportal /build_dir/swagger.json /usr/share/nginx/html

COPY make/photon/portal/nginx.conf /etc/nginx/nginx.conf

RUN tdnf install -y nginx sudo >> /dev/null \
    && ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log \
    && groupadd -r -g 10000 nginx && useradd --no-log-init -r -g 10000 -u 10000 nginx \
    && chown -R nginx:nginx /etc/nginx \
    && tdnf clean all

EXPOSE 80
VOLUME /var/cache/nginx /var/log/nginx /run

STOPSIGNAL SIGQUIT

HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080 || exit 1
USER nginx
CMD ["nginx", "-g", "daemon off;"]

