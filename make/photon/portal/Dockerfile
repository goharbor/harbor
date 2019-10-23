FROM node:10.15.0 as nodeportal

ARG npm_registry=https://registry.npmjs.org
ENV NPM_CONFIG_REGISTRY=${npm_registry}

RUN mkdir -p /portal_src
RUN mkdir -p /build_dir

COPY make/photon/portal/entrypoint.sh /
COPY src/portal /portal_src
COPY ./docs/swagger.yaml   /portal_src

WORKDIR /portal_src

RUN npm install && \
    chmod u+x /entrypoint.sh
RUN /entrypoint.sh
VOLUME ["/portal_src"]


FROM photon:2.0

RUN tdnf install -y nginx >> /dev/null \
    && ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log \
    && tdnf clean all

EXPOSE 80
VOLUME /var/cache/nginx /var/log/nginx /run


COPY --from=nodeportal /build_dir/dist /usr/share/nginx/html
COPY --from=nodeportal /build_dir/swagger.yaml /usr/share/nginx/html
COPY --from=nodeportal /build_dir/swagger.json /usr/share/nginx/html

COPY make/photon/portal/nginx.conf /etc/nginx/nginx.conf

STOPSIGNAL SIGQUIT

HEALTHCHECK CMD curl --fail -s http://127.0.0.1 || exit 1

CMD ["nginx", "-g", "daemon off;"]
