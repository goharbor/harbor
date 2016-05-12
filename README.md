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
    4, SQL_PATH (dir name contains sql files, /go/bin/sql

  * COPY /path/to/sql/dir into docker container, accessible by SQL_PATH

  * COPY /code/root/CATEGORIES into docner container, same dir as harbor
    binary

  * Nginx forward change, add harbor as upstream, forward
  /api/v3/repositories to harbor.


# 如何准备并且发布一个应用（基础镜像)
  

  基础镜像模板参见项目：
  ```
    https://github.com/Dataman-Cloud/AppCatalogTemplates
  ```

  例如，树人云发布MySQL基础镜像

  * 从官方或者自己写Dockerfile打一个MySQL镜像，
    将此镜像push到树人云registry的library namespace下(需要管理员账号)
    docker push xxxregistry.dataman-inc.com/library/mysql:v5.5

  * 编写MySQL的readme
    (templates/library/mysql/readme.md)，readme中应该包含此镜像的用法，原理等,markdown格式。

  * 编写mysql对应的docker_compose.yml(templates/library/mysql/docker_compose.yml)文件，语法参考docker compose 编写规范
  
  * 编写MySQL对应的MarathonConfig，
    规定单个container占用的cpu，mem，instances等，具体格式参考templates/library/example/marathon_config.yml

  * 提取启动MySQL需要的变量， 供用户自由填写，
    具体格式参考templates/library/example/catalog.yml

  * 准备镜像需要的mysql.png，拷贝到templates/library/mysql/mysql.png, 此图片最终显示到应用目录的列表页面。

  * 代码根目录执行./repo_icon_rsync.sh
    此命令收集所有的图片，拷贝图片到templates/static下，并rsync静态文件到服务器上，五分钟后运维自动分发。

  * 分别修改MySQL的分类， 描述，
    是否公共等信息，具体参考templates/library/example文件夹。

  * 所有材料准备完毕之后，需要修改repo_updater.sh脚本

  * 应用目录的分类信息存储在代码目录的CATEGORIES文件中，如需要修改可直接修改后提交。

  * readme的写法可参考templates/example目录下写法， 或者参考rancher
    app_catalog功能。

# docker_compose 文件支持的compose参数包括

	Image        string      `json:"image" yaml:"image"`
	Command      interface{} `json:"command" yaml:"command"`
	EntryPoint   string      `json:"entrypoint" yaml:"entrypoint"`
	Environment  Environment `json:"environment" yaml:"environment"`
	Labels       Labels      `json:"labels" yaml:"labels"`
	Volumes      []*Volume   `json:"volumes" yaml:"volumes"`
	Expose       []int       `json:"expose" yaml:"expose"`
	Ports        []*Port     `json:"ports" yaml:"ports"`
	Net          string      `json:"net" yaml:"net"`                   // bridge, host
	NetworkMode  string      `json:"network_mode" yaml:"network_mode"` //compose version2 for net, same as net
	Links       []*Link     `json:"links" yaml:"links,omitempty"`



# marathon_config 包含参数包括

	Cpu          float32     `json:"cpu" yaml:"cpu"`
	Mem          float32     `json:"mem" yaml:"mem"`
	Instances    int32       `json:"instances" yaml:"instances"`
	LogPaths     []string    `json:"log_paths" yaml:"log_paths"`

