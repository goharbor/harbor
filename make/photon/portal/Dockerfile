ARG harbor_base_image_version
ARG harbor_base_namespace
FROM node:16.18.0 as nodeportal

WORKDIR /build_dir

ARG npm_registry=https://registry.npmjs.org

COPY ./api/v2.0/swagger.yaml /build_dir/swagger.yaml
COPY src/portal /build_dir

ENV NPM_CONFIG_REGISTRY=${npm_registry}
RUN npm install --unsafe-perm
RUN npm run generate-build-timestamp
RUN node --max_old_space_size=2048 'node_modules/@angular/cli/bin/ng' build --configuration production
RUN npm install js-yaml@4.1.0 \
    && node -e "const yaml = require('js-yaml'); const fs = require('fs'); const swagger = yaml.load(fs.readFileSync('swagger.yaml', 'utf8')); fs.writeFileSync('swagger.json', JSON.stringify(swagger));"

COPY ./LICENSE /build_dir/dist

RUN cd app-swagger-ui && npm install --unsafe-perm
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

