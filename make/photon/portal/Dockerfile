ARG harbor_base_image_version
ARG harbor_base_namespace
ARG NODE
FROM ${NODE} as nodeportal

WORKDIR /build_dir

ARG npm_registry=https://registry.npmjs.org

COPY ./api/v2.0/swagger.yaml /build_dir/swagger.yaml
COPY src/portal /build_dir

ENV NPM_CONFIG_REGISTRY=${npm_registry}
ENV NODE_OPTIONS=--max_old_space_size=4096

RUN --mount=type=cache,target=/root/.npm \
    npm ci --unsafe-perm --no-optional --no-audit --no-fund --loglevel=error

RUN npm run generate-build-timestamp
RUN node --max_old_space_size=2048 'node_modules/@angular/cli/bin/ng' build --configuration production
RUN node -e "const yaml = require('js-yaml'); const fs = require('fs'); const swagger = yaml.load(fs.readFileSync('swagger.yaml', 'utf8')); fs.writeFileSync('swagger.json', JSON.stringify(swagger));"

COPY ./LICENSE /build_dir/dist

RUN --mount=type=cache,target=/root/.npm \
    cd app-swagger-ui && npm ci --unsafe-perm --no-optional --no-audit --no-fund --loglevel=error
RUN cd app-swagger-ui && npm run build

FROM ${harbor_base_namespace}/harbor-portal-base:${harbor_base_image_version}

COPY --from=nodeportal /build_dir/dist /usr/share/nginx/html
COPY --from=nodeportal /build_dir/swagger.json /usr/share/nginx/html
COPY --from=nodeportal /build_dir/app-swagger-ui/dist /usr/share/nginx/html

VOLUME /var/cache/nginx /var/log/nginx /run

STOPSIGNAL SIGQUIT

HEALTHCHECK CMD curl --fail -s http://localhost:8080 || curl -k --fail -s https://localhost:8443 || exit 1
USER nginx
CMD ["nginx", "-g", "daemon off;"]

