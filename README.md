# Harbor

Project Harbor is modified version of Vmware Harbor project, with
features added to facility Shurenyun internal usage.

# Features Added

  * Repository and tag info collected when docker push/pull
  * API support repository and tag info retriving and updating
  * Shurenyun Compose support - which support auto App generations
  * mysql auto updating support now

# Deploy(After Modification)

  * new environments variable added

    1, REDIS_HOST
    2, REDIS_PORT
    3, APP_API_URL(foward nginx url, https://forward.dataman-inc.net)
    4, SQL_PATH (dir name contains sql files, /go/bin/sql

  * COPY /path/to/sql/dir into docker container, accessible by SQL_PATH

  * Nginx forward change, add harbor as upstream, forward
  /api/v3/repositories to harbor.


