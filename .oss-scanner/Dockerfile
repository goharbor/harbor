# goharbor/harbor: Go services (core, jobservice, registryctl, exporter, db migrator) and the Angular portal.
FROM node:22.22.3-bookworm AS node

FROM golang:1.27.2-bookworm
COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/lib/node_modules /usr/local/lib/node_modules
RUN ln -sf /usr/local/lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm \
    && ln -sf /usr/local/lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx

# PostgreSQL 18 matches make/photon/db; Redis stands in for Valkey. Several tests hardcode redis://localhost:6379.
# Chromium runs the karma portal tests.
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl git make gcc libc6-dev \
      postgresql-common redis-server chromium \
    && /usr/share/postgresql-common/pgdg/apt.postgresql.org.sh -y \
    && apt-get install -y --no-install-recommends postgresql-18 \
    && rm -rf /var/lib/apt/lists/*

RUN pg_ctlcluster 18 main start \
    && su postgres -c "psql -c \"ALTER USER postgres PASSWORD 'root123'\"" \
    && su postgres -c "createdb registry" \
    && pg_ctlcluster 18 main stop
RUN printf '%s\n' '#!/bin/sh' 'set -e' \
      'pg_ctlcluster 18 main status >/dev/null 2>&1 || pg_ctlcluster 18 main start' \
      'redis-cli ping >/dev/null 2>&1 || redis-server --daemonize yes --port 6379 --save "" --appendonly no' \
      'timeout 60 sh -c "until pg_isready -h localhost -q; do sleep 0.5; done"' \
      'timeout 30 sh -c "until redis-cli ping >/dev/null 2>&1; do sleep 0.5; done"' \
      'echo "postgres on localhost:5432 (postgres/root123, db registry), redis on localhost:6379"' \
      > /usr/local/bin/harbor-test-services \
    && chmod +x /usr/local/bin/harbor-test-services

# Same variables as .github/workflows/CI.yml, so DB tests work after `harbor-test-services`.
ENV POSTGRESQL_HOST=localhost \
    POSTGRESQL_PORT=5432 \
    POSTGRESQL_USR=postgres \
    POSTGRESQL_PWD=root123 \
    POSTGRESQL_DATABASE=registry \
    POSTGRES_MIGRATION_SCRIPTS_PATH=/src/make/migrations/postgresql \
    REDIS_HOST=localhost \
    HARBOR_ADMIN=admin \
    HARBOR_ADMIN_PASSWD=Harbor12345 \
    CORE_SECRET=tempString \
    UTTEST=true \
    GOFLAGS=-buildvcs=false \
    CHROME_BIN=/usr/bin/chromium \
    TOKEN_PRIVATE_KEY_PATH=/src/tests/private_key.pem

RUN go install github.com/go-swagger/go-swagger/cmd/swagger@v0.33.1

# scanner contract: the checkout lives inside the image, at /src
COPY . /src
WORKDIR /src

# Same arguments as `make gen_apis`; the server code under src/server/v2.0 is generated, not committed.
RUN swagger generate server --template-dir=./tools/swagger/templates --exclude-main \
      --additional-initialism=CVE --additional-initialism=GC --additional-initialism=OIDC \
      -f ./api/v2.0/swagger.yaml -A harbor --target ./src/server/v2.0

WORKDIR /src/src
RUN go mod download
RUN go build -o /usr/local/bin/harbor-core ./core \
    && go build -o /usr/local/bin/harbor-jobservice ./jobservice \
    && go build -o /usr/local/bin/harbor-registryctl ./registryctl \
    && go build -o /usr/local/bin/harbor-exporter ./cmd/exporter \
    && go build -o /usr/local/bin/harbor-db-migrator ./cmd/standalone-db-migrator

# Compile every test binary while online so `go test` works offline. -exec /bin/true skips TestMain, which needs
# PostgreSQL; running the tests needs harbor-test-services.
RUN go test -count=1 -exec /bin/true ./... >/dev/null

# Same steps as make/photon/portal/Dockerfile; postinstall generates the TypeScript API client from swagger.yaml.
WORKDIR /src/src/portal
ENV NODE_OPTIONS=--max_old_space_size=4096
RUN npm ci --no-audit --no-fund \
    && npm run generate-build-timestamp \
    && node --max_old_space_size=2048 node_modules/@angular/cli/bin/ng build --configuration production
WORKDIR /src/src/portal/app-swagger-ui
RUN npm ci --no-audit --no-fund && npm run build

WORKDIR /src
