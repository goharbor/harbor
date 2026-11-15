# Drives the API and reads pg_stat_activity: curl for one, psql for the other.
FROM docker.io/library/alpine:3.22
RUN apk add --no-cache bash curl postgresql16-client jq
