# Harbor

Project Harbor is modified version of Vmware Harbor project, with
features added to facility Shurenyun internal usage.

# Features Added

  * Repository and tag info collected when docker push/pull
  * API support repository and tag info retriving and updating
  * Shurenyun Compose support - which support auto app generation
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


# 如何准备并且发布一个应用（基础镜像)

  例如，树人云发布mysql基础镜像

  * 从官方或者自己写Dockerfile打一个MySQL镜像，
    将此镜像push到树人云registry的library namespace下(需要管理员账号)
    docker push xxxregistry.dataman-inc.com/library/mysql:v5.5

  * 编写MySQL的readme
    (templates/library/mysql/readme.md)，readme中应该包含此镜像的用法，原理等,markdown格式。

  * 编写mysql对应的sry_compose.yml(templates/library/mysql/sry_compose.yml)文件，把可变的变量提取到questions部分当中，已备用户填写或者选择。

  * 准备镜像需要的mysql.png，拷贝到templates/library/mysql/mysql.png, 此图片最终显示到应用目录的列表页面。

  * 代码根目录执行./repo_icon_rsync.sh
    此命令收集所有的图片，拷贝图片到templates/static下，并rsync静态文件到服务器上，五分钟后运维自动分发。

  * 所有材料准备完毕之后，需要修改repo_updater.sh脚本，此脚本目的是修改repo的描述，分类和是否发布，详情可参看脚本内部。

  * 应用目录的分类信息存储在代码目录的CATEGORIES文件中，如需要修改可直接修改后提交。

  * readme的写法可参考templates/example目录下写法， 或者参考rancher
    app_catalog功能。
