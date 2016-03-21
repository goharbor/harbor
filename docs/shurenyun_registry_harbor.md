#shurenyun提供的registry
    文档信息
    创建人 韩路
    邮件地址 lhan@dataman-inc.com
    建立时间 2016年3月17号

##流程图
![harbor](imp/harbor.png)
docker向registry请求，registry先去向harbor进行认证，harbor通过对mysql中的数据匹配，完成认证．

##1. registry线下版
####1.1 目录结构
    registry/
        |---config/
        |   |---registry/
        |   |   |---config.yml
        |   |   |---root.crt
        |---certs/
        |   |---shurenyun_com.key
        |   |---shurenyun_com.crt
        |---registry.sh


####1.2 registry.sh
```
docker run -d -v ./config/registry/:/etc/registry/ \
              -v ./certs:/certs \
              -v /data/registry:/storage \
              -e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/shurenyun_com.crt \
              -e REGISTRY_HTTP_TLS_KEY=/certs/shurenyun_com.key \
              -p 443:5000 \
              registry:2.3.0 /etc/registry/config.yml
```
####1.3 config.yml
```
version: 0.1
log:
  level: debug
  fields:
    service: registry
storage:
    cache:
        layerinfo: inmemory
    filesystem:
        rootdirectory: /storage
    maintenance:
        uploadpurging:
            enabled: false
http:
    addr: :5000
    secret: placeholder
auth:
  token:
    issuer: registry-token-issuer
    realm: https://index.shurenyun.com/service/token
    rootcertbundle: /etc/registry/root.crt
    service: token-service

notifications:
  endpoints:
      - name: harbor
        disabled: false
        url: https://index.shurenyun.com/service/notifications
        timeout: 500
        threshold: 5
        backoff: 1000

```

##2. harbor UI
####2.1 目录结构
    harbor_ui/
        |---config/
        |   |---ui/
        |   |   |---app.conf
        |   |   |---private_key.pem
        |---harbor_ui.sh


####2.2 harbor_ui.sh
```
docker run -d -v ./config/ui/app.conf:/etc/ui/app.conf \
              -v ./config/ui/private_key.pem:/etc/ui/private_key.pem \
              -e MYSQL_HOST=mysql \
              -e MYSQL_PORT_3306_TCP_PORT=3306 \
              -e MYSQL_USR=root \
              -e MYSQL_PWD=111111 \
              -e REGISTRY_URL=http://{registry}:443 \
              -e CONFIG_PATH=/etc/ui/app.conf \
              -e HARBOR_REG_URL=https://{registry}\signUp \
              -e HARBOR_ADMIN_PASSWORD=Harbor12345 \
              -e HARBOR_URL=https://{registry} \
              -e AUTH_MODE=db_auth \
              -e LDAP_URL=ldaps://ldap.mydomain.com \
              -e LDAP_BASE_DN=uid=%s,ou=people,dc=mydomain,dc=com \
              -p 80:80
              registry.shurenyun.com/harbor_ui:v1
```

####2.3 app.conf
```
appname = registry
runmode = dev

[lang]
types = en-US|zh-CN
names = en-US|zh-CN

[dev]
httpport = 80

[mail]
host = smtp.mydomain.com
port = 25
username = sample_admin@mydomain.com
password = abc
from = admin <sample_admin@mydomain.com>
```

##3. mysql
####3.1 sql文件
```
drop database if exists registry;
create database registry charset = utf8;

use registry;

create table access (
 access_id int NOT NULL AUTO_INCREMENT,
 access_code char(1),
 comment varchar (30),
 primary key (access_id)
);

insert into access values
( null, 'A', 'All access for the system'),
( null, 'M', 'Management access for project'),
( null, 'R', 'Read access for project'),
( null, 'W', 'Write access for project'),
( null, 'D', 'Delete access for project'),
( null, 'S', 'Search access for project');


create table role (
 role_id int NOT NULL AUTO_INCREMENT,
 role_code varchar(20),
 name varchar (20),
 primary key (role_id)
);

insert into role values
( null, 'AMDRWS', 'sysAdmin'),
( null, 'MDRWS', 'projectAdmin'),
( null, 'RWS', 'developer'),
( null, 'RS', 'guest');


create table user (
 user_id int NOT NULL AUTO_INCREMENT,
 username varchar(15),
 email varchar(30),
 password varchar(40) NOT NULL,
 realname varchar (20) NOT NULL,
 comment varchar (30),
 deleted tinyint (1) DEFAULT 0 NOT NULL,
 reset_uuid varchar(40) DEFAULT NULL,
 salt varchar(40) DEFAULT NULL,
 primary key (user_id),
 UNIQUE (username),
 UNIQUE (email)
);

insert into user values
(1, 'admin', 'admin@example.com', '', 'system admin', 'admin user',0, null, ''),
(2, 'anonymous', 'anonymous@example.com', '', 'anonymous user', 'anonymous user', 1, null, '');

create table project (
 project_id int NOT NULL AUTO_INCREMENT,
 owner_id int NOT NULL,
 name varchar (30) NOT NULL,
 creation_time timestamp,
 deleted tinyint (1) DEFAULT 0 NOT NULL,
 public tinyint (1) DEFAULT 0 NOT NULL,
 primary key (project_id),
 FOREIGN KEY (owner_id) REFERENCES user(user_id),
 UNIQUE (name)
);

insert into project values
(null, 1, 'library', NOW(), 0, 1);

create table project_role (
 pr_id int NOT NULL AUTO_INCREMENT,
 project_id int NOT NULL,
 role_id int NOT NULL,
 primary key (pr_id),
 FOREIGN KEY (role_id) REFERENCES role(role_id),
 FOREIGN KEY (project_id) REFERENCES project (project_id)
);

insert into project_role values
( 1,1,1 );

create table user_project_role (
 upr_id int NOT NULL AUTO_INCREMENT,
 user_id int NOT NULL,
 pr_id int NOT NULL,
 primary key (upr_id),
 FOREIGN KEY (user_id) REFERENCES user(user_id),
 FOREIGN KEY (pr_id) REFERENCES project_role (pr_id)
);

insert into user_project_role values
( 1,1,1 );

create table access_log (
 log_id int NOT NULL AUTO_INCREMENT,
 user_id int NOT NULL,
 project_id int NOT NULL,
 repo_name varchar (40),
 GUID varchar(64),
 operation varchar(20) NOT NULL,
 op_time timestamp,
 primary key (log_id),
 FOREIGN KEY (user_id) REFERENCES user(user_id),
 FOREIGN KEY (project_id) REFERENCES project (project_id)
);
```

####3.2 mysql_run.sh
`docker run --name some-mysql -e MYSQL_ROOT_PASSWORD=111111 -p 3306:3306 -d mysql:5.6`


## 参考文档:
* https://github.com/vmware/harbor


