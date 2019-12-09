ARG harbor_base_image_version
FROM node:10.15.0 as nodeportal

WORKDIR /build_dir

ARG npm_registry=https://registry.npmjs.org
ENV NPM_CONFIG_REGISTRY=${npm_registry}

COPY src/portal/package.json /build_dir
COPY src/portal/package-lock.json /build_dir
COPY ./api/harbor/swagger.yaml /build_dir

RUN apt-get update \
    && apt-get install -y --no-install-recommends python-yaml=3.12-1 \
    && python -c 'import sys, yaml, json; y=yaml.load(sys.stdin.read()); print json.dumps(y)' < swagger.yaml > swagger.json \
    && npm install

COPY ./LICENSE /build_dir
COPY src/portal /build_dir

RUN ls -la \
    && npm run build_lib \
    && npm run link_lib \
    && node --max_old_space_size=2048 'node_modules/@angular/cli/bin/ng' build --prod

FROM goharbor/harbor-portal-base:${harbor_base_image_version}

COPY --from=nodeportal /build_dir/dist /usr/share/nginx/html
COPY --from=nodeportal /build_dir/swagger.yaml /usr/share/nginx/html
COPY --from=nodeportal /build_dir/swagger.json /usr/share/nginx/html
COPY --from=nodeportal /build_dir/LICENSE /usr/share/nginx/html

COPY make/photon/portal/nginx.conf /etc/nginx/nginx.conf

EXPOSE 8080
VOLUME /var/cache/nginx /var/log/nginx /run

STOPSIGNAL SIGQUIT

HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080 || exit 1
USER nginx
CMD ["nginx", "-g", "daemon off;"]

