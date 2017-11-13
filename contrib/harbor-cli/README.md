## About This Project

Project Harbor is an enterprise-class registry server that stores and distributes Docker images. Harbor extends the open source Docker Distribution by adding the functionalities usually required by an enterprise, such as security, identity and management. As an enterprise private registry, Harbor offers better performance and security. Having a registry closer to the build and run environment improves the image transfer efficiency. Harbor supports the setup of multiple registries and has images replicated between them. With Harbor, the images are stored within the private registry, keeping the bits and intellectual properties behind the company firewall. In addition, Harbor offers advanced security features, such as user management, access control and activity auditing.

This project provides a great native command-line experience for managing Harbor resources like user, project, etc. It can be used on macOS, Linux, and Docker.

## Install Harbor CLI

Harbor CLI can be installed by one of two approaches:

* Option 1: Build as a Docker image(easy, recommended)
* Option 2: Native Installation from Source
* Option 3: Install from pypi

### Option 1: Build as a Docker image(easy, recommended)

We maintain a Docker prebuilt image with Harbor CLI. Install the CLI using `docker run`.

```sh
docker run -t -i krystism/harborclient harbor help
```

We strongly suggest you build image from code manually, because our prebuilt image may be not latest version.

```sh
docker build -t yourname/harborclient .
```

Run Harbor CLI as follows:

```bash
$ docker run --rm \
 -e HARBOR_USERNAME="admin" \
 -e HARBOR_PASSWORD="Harbor12345" \
 -e HARBOR_PROJECT=1 \
 -e HARBOR_URL="http://localhost" \
 yourname/harborclient harbor info

+------------------------------+---------------------+
| Property                     | Value               |
+------------------------------+---------------------+
| admiral_endpoint             | NA                  |
| auth_mode                    | db_auth             |
| disk_free                    | 4993355776          |
| disk_total                   | 18381979648         |
| harbor_version               | v1.2.2              |
| has_ca_root                  | False               |
| next_scan_all                | 0                   |
| project_creation_restriction | everyone            |
| registry_url                 | localhost           |
| self_registration            | True                |
| with_admiral                 | False               |
| with_clair                   | False               |
| with_notary                  | False               |
+------------------------------+---------------------+
```

Create an alias:

```bash
alias harbor='docker run \
 -e HARBOR_USERNAME="admin" \
 -e HARBOR_PASSWORD="Harbor12345" \
 -e HARBOR_URL="http://localhost" \
 --rm krystism/harborclient harbor'
```

Then you can run Harbor CLI like:

```
$ harbor user-list
+---------+----------+----------+----------------------+--------------+-------------+
| user_id | username | is_admin |        email         |   realname   |   comment   |
+---------+----------+----------+----------------------+--------------+-------------+
|    1    |  admin   |    1     |  admin@example.com   | system admin |  admin user |
|    2    | int32bit |    0     | int32bit@example.com |   int32bit   |  int32bit   |
+---------+----------+----------+----------------------+--------------+-------------+
```

### Option 2: Native Installation from Source

The installation steps boil down to the following:

#### Install requirements

```
sudo pip install -r requirements.txt
```

#### Install Harbor CLI.

```sh
sudo python setup.py install
```

Or

```sh
sudo pip install .
```

### Option 3: Install from pypi

```
sudo pip install python-harborclient
```

### Verify operation

As the `admin` user, do a `info` request:

```
$ harbor --os-baseurl https://localhost --os-username admin --os-project 1 info
password: *****
+------------------------------+---------------------+
| Property                     | Value               |
+------------------------------+---------------------+
| admiral_endpoint             | NA                  |
| auth_mode                    | db_auth             |
| disk_free                    | 4992696320          |
| disk_total                   | 18381979648         |
| harbor_version               | v1.2.2              |
| has_ca_root                  | False               |
| next_scan_all                | 0                   |
| project_creation_restriction | everyone            |
| registry_url                 | localhost           |
| self_registration            | True                |
| with_admiral                 | False               |
| with_clair                   | False               |
| with_notary                  | False               |
+------------------------------+---------------------+
```

### Create harbor client environment scripts

To increase efficiency of client operations, Harbor CLI supports simple client environment scrips also known as `harborrc` file.
These scripts typically contain common options for all client, but also support unique options.

Create client environment scripts for `admin` user:

```bash
cat >admin-harborrc <<EOF
export HARBOR_USERNAME=admin
export HARBOR_PASSWORD=Harbor12345
export HARBOR_URL=http://localhost
export HARBOR_PROJECT=1
EOF
```

Replace `HARBOR_PASSWORD` with your password.

To run clients as a specific project and user, you can simply load the associated client environment script prior to running them.

```bash
source admin-harborrc
```

List images:

```bash
$ harbor list
+-----------------------+------------+-----------+------------+------------+------------+----------------------+
|          name         | project_id |    size   | tags_count | star_count | pull_count |     update_time      |
+-----------------------+------------+-----------+------------+------------+------------+----------------------+
|    int32bit/busybox   |     2      |   715181  |     1      |     0      |     0      | 2017-11-01T07:06:36Z |
| int32bit/golang:1.7.3 |     2      | 257883053 |     2      |     0      |     0      | 2017-11-01T12:59:05Z |
|  int32bit/hello-world |     2      |    974    |     1      |     0      |     0      | 2017-11-01T13:22:46Z |
+-----------------------+------------+-----------+------------+------------+------------+----------------------+
```

### Setup bash completion

```bash
$ complete -W $(harbor bash-completion) harbor
$ harbor us<tab><tab>
usage user-create user-delete user-list user-show  user-update
```

## User Guide

This guide walks you through the fundamentals of using Harbor CLI. You'll learn how to use Harbor CLI to:

* Manage your projects.
* Manage members of a project.
* Search projects and repositories.
* Manage users.
* Manage replication policies.
* Manage configuration.
* Delete repositories and images.
* Show logs.
* Get statistics data.
* ...

Once you install Harbor CLI, you can run `harbor help` to get usage:

```bash
$ harbor help
usage: harbor [--debug] [--timings] [--version] [--os-username <username>]
              [--os-password <password>] [--os-project <project>]
              [--timeout <timeout>] [--os-baseurl <baseurl>] [--insecure]
              [--os-cacert <ca-certificate>] [--os-api-version <api-version>]
              <subcommand> ...
```

Run "harbor help COMMAND" for help on a specific command.

```bash
$ harbor help user-create
usage: harbor user-create --username <username> --password <password> --email
                          <email> [--realname <realname>]
                          [--comment <comment>]

Create a new User.

Optional arguments:
  --username <username>  Unique name of the new user.
  --password <password>  Password of the new user.
  --email <email>        Email of the new user.
  --realname <realname>  Realname of the new user.
  --comment <comment>    Comment of the new user.
```

Show details about API requests using `--debug` option:

```bash
$ harbor  --debug --insecure project-list
DEBUG (connectionpool:824) Starting new HTTPS connection (1): devstack
DEBUG (connectionpool:396) https://devstack:443 "POST /login HTTP/1.1" 200 0
DEBUG (client:274) Successfully login, session id: 2642a18db2cb0fb207bd721899da9f8b
REQ: curl -g -i --insecure 'https://devstack/api/projects' -X GET -H "Accept: application/json" -H "Harbor-API-Version: v2" -H "User-Agent: python-harborclient" -b "beegosessionID: 2642a18db2cb0fb207bd721899da9f8b"
DEBUG (connectionpool:824) Starting new HTTPS connection (1): devstack
DEBUG (connectionpool:396) https://devstack:443 "GET /api/projects HTTP/1.1" 200 316
RESP: [200] {'Content-Length': '316', 'Content-Encoding': 'gzip', 'X-Total-Count': '2', 'Server': 'nginx/1.11.13', 'Connection': 'keep-alive', 'Date': 'Mon, 06 Nov 2017 12:24:53 GMT', 'Content-Type': 'application/json; charset=utf-8'}
RESP BODY: [{"creation_time_str": "", "enable_content_trust": false, "Togglable": true, "owner_name": "", "name": "int32bit", "deleted": 0, "repo_count": 3, "creation_time": "2017-11-01T06:56:07Z", "update_time": "2017-11-01T06:56:07Z", "prevent_vulnerable_images_from_running": false, "current_user_role_id": 1, "project_id": 2, "automatically_scan_images_on_push": false, "public": 1, "prevent_vulnerable_images_from_running_severity": "", "owner_id": 1}, {"creation_time_str": "", "enable_content_trust": false, "Togglable": true, "owner_name": "", "name": "library", "deleted": 0, "repo_count": 0, "creation_time": "2017-11-01T06:08:43Z", "update_time": "2017-11-01T06:08:43Z", "prevent_vulnerable_images_from_running": false, "current_user_role_id": 1, "project_id": 1, "automatically_scan_images_on_push": false, "public": 1, "prevent_vulnerable_images_from_running_severity": "", "owner_id": 1}]

+------------+----------+----------+----------------------+------------+----------------------+--------+
| project_id |   name   | owner_id | current_user_role_id | repo_count |    creation_time     | public |
+------------+----------+----------+----------------------+------------+----------------------+--------+
|     1      | library  |    1     |          1           |     0      | 2017-11-01T06:08:43Z |   1    |
|     2      | int32bit |    1     |          1           |     3      | 2017-11-01T06:56:07Z |   1    |
+------------+----------+----------+----------------------+------------+----------------------+--------+
```

Print call timing info with `--timings` option:

```
$ harbor  --insecure --timings user-list
+---------+----------+----------+----------------------+--------------+-------------+
| user_id | username | is_admin |        email         |   realname   |   comment   |
+---------+----------+----------+----------------------+--------------+-------------+
|    1    |  admin   |    1     |  admin@example.com   | system admin |  admin user |
|    3    | int32bit |    0     | int32bit@example.com |   int32bit   |     test    |
+---------+----------+----------+----------------------+--------------+-------------+
+--------------+-----------------+
| url          | seconds         |
+--------------+-----------------+
| GET /users   | 0.0146510601044 |
| GET /users/1 | 0.0146780014038 |
| Total        | 0.0293290615082 |
+--------------+-----------------+
Total: 0.0293290615082 seconds
```

All SSL connections are attempted to be made secure by using the CA certificate bundle installed by default. This makes all connections considered "insecure" fail unless `--insecure` is used.

```
$ harbor info
Traceback (most recent call last):
  File "/usr/local/bin/harbor", line 10, in <module>
    sys.exit(main())
  File "/usr/local/lib/python2.7/dist-packages/harborclient/shell.py", line 404, in main
    HarborShell().main(argv)
  File "/usr/local/lib/python2.7/dist-packages/harborclient/shell.py", line 330, in main
    self.cs.authenticate()
  File "/usr/local/lib/python2.7/dist-packages/harborclient/v2/client.py", line 83, in authenticate
    self.client.authenticate()
  File "/usr/local/lib/python2.7/dist-packages/harborclient/client.py", line 270, in authenticate
    verify=self.verify_cert)
  File "/usr/local/lib/python2.7/dist-packages/requests/api.py", line 112, in post
    return request('post', url, data=data, json=json, **kwargs)
  File "/usr/local/lib/python2.7/dist-packages/requests/api.py", line 58, in request
    return session.request(method=method, url=url, **kwargs)
  File "/usr/local/lib/python2.7/dist-packages/requests/sessions.py", line 508, in request
    resp = self.send(prep, **send_kwargs)
  File "/usr/local/lib/python2.7/dist-packages/requests/sessions.py", line 618, in send
    r = adapter.send(request, **kwargs)
  File "/usr/local/lib/python2.7/dist-packages/requests/adapters.py", line 506, in send
    raise SSLError(e, request=request)
requests.exceptions.SSLError: HTTPSConnectionPool(host='devstack', port=443): Max retries exceeded with url: /login (Caused by SSLError(SSLError("bad handshake: Error([('SSL routines', 'tls_process_server_certificate', 'certificate verify failed')],)",),))
$ harbor --insecure info
+------------------------------+---------------------+
| Property                     | Value               |
+------------------------------+---------------------+
| admiral_endpoint             | NA                  |
| auth_mode                    | db_auth             |
| disk_free                    | 4991021056          |
| disk_total                   | 18381979648         |
| harbor_version               | v1.2.2              |
| has_ca_root                  | False               |
| next_scan_all                | 0                   |
| project_creation_restriction | everyone            |
| registry_url                 | 192.168.99.101:8888 |
| self_registration            | True                |
| with_admiral                 | False               |
| with_clair                   | False               |
| with_notary                  | False               |
+------------------------------+---------------------+
```

## Examples

### Create a new user

```
$ harbor --insecure user-create \
 --username new-user \
 --password 1q2w3e4r \
 --email new_user@example.com \
 --realname newuser \
 --comment "I am a new user"
Create user 'new-user' successfully.
```

### Delete a user

```
$ harbor --insecure user-delete new-user
Delete user 'new-user' sucessfully.
```

### List repositories and images

```
$ harbor  list
+-----------------------+------------+-----------+------------+------------+------------+----------------------+
|          name         | project_id |    size   | tags_count | star_count | pull_count |     update_time      |
+-----------------------+------------+-----------+------------+------------+------------+----------------------+
|    int32bit/busybox   |     2      |   715181  |     1      |     0      |     0      | 2017-11-01T07:06:36Z |
| int32bit/golang:1.7.3 |     2      | 257883053 |     2      |     0      |     0      | 2017-11-01T12:59:05Z |
|  int32bit/hello-world |     2      |    974    |     1      |     0      |     0      | 2017-11-01T13:22:46Z |
+-----------------------+------------+-----------+------------+------------+------------+----------------------+
```

### Show details about image

```
$ harbor  show int32bit/golang:1.7.3
+--------------------+-------------------------------------------------------------------------+
| Property           | Value                                                                   |
+--------------------+-------------------------------------------------------------------------+
| creation_time      | 2017-11-01T12:59:05Z                                                    |
| description        |                                                                         |
| id                 | 2                                                                       |
| name               | int32bit/golang                                                         |
| project_id         | 2                                                                       |
| pull_count         | 0                                                                       |
| star_count         | 0                                                                       |
| tag_architecture   | amd64                                                                   |
| tag_author         |                                                                         |
| tag_created        | 2016-11-08T19:32:39.908048617Z                                          |
| tag_digest         | sha256:37d263ccd240e113a752c46306ad004e36532ce118eb3131d9f76f43cc606d5d |
| tag_docker_version | 1.12.3                                                                  |
| tag_name           | 1.7.3                                                                   |
| tag_os             | linux                                                                   |
| tag_signature      | -                                                                       |
| tags_count         | 2                                                                       |
| update_time        | 2017-11-01T12:59:05Z                                                    |
+--------------------+-------------------------------------------------------------------------+
```

### Get top accessed repositories

```
$ harbor top
+----------------------+------------+------------+
|         name         | pull_count | star_count |
+----------------------+------------+------------+
|   int32bit/busybox   |     10     |     0      |
|   int32bit/golang    |     8      |     0      |
| int32bit/hello-world |     1      |     0      |
+----------------------+------------+------------+
```

### Lists members of a project.

```
$ harbor member-list
+----------+--------------+---------+---------+
| username |  role_name   | user_id | role_id |
+----------+--------------+---------+---------+
|  admin   | projectAdmin |    1    |    1    |
|   foo    |  developer   |    5    |    2    |
|   test   |    guest     |    6    |    3    |
+----------+--------------+---------+---------+
```

### Show logs

```
$ harbor logs
+--------+----------------------+----------+------------+-----------+-----------------------------+
| log_id |       op_time        | username | project_id | operation |          repository         |
+--------+----------------------+----------+------------+-----------+-----------------------------+
|   1    | 2017-11-01T06:56:07Z |  admin   |     2      |   create  |          int32bit/          |
|   2    | 2017-11-01T07:06:36Z |  admin   |     2      |    push   |   int32bit/busybox:latest   |
|   3    | 2017-11-01T12:59:05Z |  admin   |     2      |    push   |    int32bit/golang:1.7.3    |
|   4    | 2017-11-01T13:22:46Z |  admin   |     2      |    push   | int32bit/hello-world:latest |
|   5    | 2017-11-01T14:21:49Z |  admin   |     2      |    push   |    int32bit/golang:latest   |
|   6    | 2017-11-03T20:39:04Z |  admin   |     3      |   create  |            test/            |
|   7    | 2017-11-03T20:39:22Z |  admin   |     3      |   delete  |            test/            |
|   8    | 2017-11-03T20:39:38Z |  admin   |     4      |   create  |            test/            |
|   9    | 2017-11-03T20:49:33Z |  admin   |     4      |   delete  |            test/            |
+--------+----------------------+----------+------------+-----------+-----------------------------+
```

### Search projects and repositories.

```
$ harbor search int32bit
Find 1 Projects:
+------------+----------+--------+------------+----------------------+
| project_id |   name   | public | repo_count |    creation_time     |
+------------+----------+--------+------------+----------------------+
|     2      | int32bit |   1    |     3      | 2017-11-01T06:56:07Z |
+------------+----------+--------+------------+----------------------+

Find 3 Repositories:
+----------------------+--------------+------------+----------------+
|   repository_name    | project_name | project_id | project_public |
+----------------------+--------------+------------+----------------+
|   int32bit/busybox   |   int32bit   |     2      |       1        |
|   int32bit/golang    |   int32bit   |     2      |       1        |
| int32bit/hello-world |   int32bit   |     2      |       1        |
+----------------------+--------------+------------+----------------+
```

### Lists targets

```
$ harbor target-list
+----+----------------------+-------------------------------------+----------+----------+----------------------+
| id |         name         |               endpoint              | username | password |    creation_time     |
+----+----------------------+-------------------------------------+----------+----------+----------------------+
| 1  |     test-target      |      http://192.168.99.101:8888     |  admin   |    -     | 2017-11-02T01:35:30Z |
| 2  |    test-target-2     |      http://192.168.99.101:9999     |  admin   |    -     | 2017-11-02T13:43:07Z |
| 3  | int32bit-test-target | http://192.168.99.101:8888/int32bit |  admin   |    -     | 2017-11-02T14:28:54Z |
+----+----------------------+-------------------------------------+----------+----------+----------------------+
```

### Ping a target

```
$ harbor target-ping 1
OK
```

### Lists replication job

```
$ harbor  job-list 1
+----+----------------------+-----------+----------+----------------------+
| id |      repository      | operation |  status  |     update_time      |
+----+----------------------+-----------+----------+----------------------+
| 1  |   int32bit/busybox   |  transfer | finished | 2017-11-02T01:35:31Z |
| 2  |   int32bit/golang    |  transfer | finished | 2017-11-02T01:35:31Z |
| 3  | int32bit/hello-world |  transfer | finished | 2017-11-02T01:35:31Z |
+----+----------------------+-----------+----------+----------------------+
```

### Show job logs:

```
$ harbor job-log  1
2017-11-02T01:35:30Z [INFO] initializing: repository: int32bit/busybox, tags: [], source URL: http://registry:5000, destination URL: http://192.168.99.101:8888, insecure: false, destination user: admin
2017-11-02T01:35:30Z [INFO] initialization completed: project: int32bit, repository: int32bit/busybox, tags: [latest], source URL: http://registry:5000, destination URL: http://192.168.99.101:8888, insecure: false, destination user: admin
2017-11-02T01:35:30Z [WARNING] the status code is 409 when creating project int32bit on http://192.168.99.101:8888 with user admin, try to do next step
2017-11-02T01:35:30Z [INFO] manifest of int32bit/busybox:latest pulled successfully from http://registry:5000: sha256:030fcb92e1487b18c974784dcc110a93147c9fc402188370fbfd17efabffc6af
2017-11-02T01:35:30Z [INFO] all blobs of int32bit/busybox:latest from http://registry:5000: [sha256:54511612f1c4d97e93430fc3d5dc2f05dfbe8fb7e6259b7351deeca95eaf2971 sha256:03b1be98f3f9b05cb57782a3a71a44aaf6ec695de5f4f8e6c1058cd42f04953e]
2017-11-02T01:35:31Z [INFO] blob sha256:54511612f1c4d97e93430fc3d5dc2f05dfbe8fb7e6259b7351deeca95eaf2971 of int32bit/busybox:latest already exists in http://192.168.99.101:8888
2017-11-02T01:35:31Z [INFO] blob sha256:03b1be98f3f9b05cb57782a3a71a44aaf6ec695de5f4f8e6c1058cd42f04953e of int32bit/busybox:latest already exists in http://192.168.99.101:8888
2017-11-02T01:35:31Z [INFO] blobs of int32bit/busybox:latest need to be transferred to http://192.168.99.101:8888: []
2017-11-02T01:35:31Z [INFO] manifest of int32bit/busybox:latest exists on source registry http://registry:5000, continue manifest pushing
2017-11-02T01:35:31Z [INFO] manifest of int32bit/busybox:latest exists on destination registry http://192.168.99.101:8888, skip manifest pushing
2017-11-02T01:35:31Z [INFO] no tag needs to be replicated, next state is "finished"
```

### Show usage

```
$ harbor usage
+-----------------------+-------+
| Property              | Value |
+-----------------------+-------+
| private_project_count | 0     |
| private_repo_count    | 0     |
| public_project_count  | 2     |
| public_repo_count     | 3     |
| total_project_count   | 2     |
| total_repo_count      | 3     |
+-----------------------+-------+
```

### Show Harbor info

```
$ harbor  info
+------------------------------+---------------------+
| Property                     | Value               |
+------------------------------+---------------------+
| admiral_endpoint             | NA                  |
| auth_mode                    | db_auth             |
| disk_free                    | 4989370368          |
| disk_total                   | 18381979648         |
| harbor_version               | v1.2.2              |
| has_ca_root                  | False               |
| next_scan_all                | 0                   |
| project_creation_restriction | everyone            |
| registry_url                 | 192.168.99.101:8888 |
| self_registration            | True                |
| with_admiral                 | False               |
| with_clair                   | False               |
| with_notary                  | False               |
+------------------------------+---------------------+
```

### Get configrations

```
$ harbor get-conf
+------------------------------+-------------------------------------------------------+----------+
|             name             |                         value                         | editable |
+------------------------------+-------------------------------------------------------+----------+
|          auth_mode           |                        db_auth                        |  False   |
|          email_from          |           admin <sample_admin@mydomain.com>           |   True   |
|          email_host          |                   smtp.mydomain.com                   |   True   |
|        email_identity        |                           -                           |   True   |
|          email_port          |                           25                          |   True   |
|          email_ssl           |                         False                         |   True   |
|        email_username        |               sample_admin@mydomain.com               |   True   |
|         ldap_base_dn         |              ou=people,dc=mydomain,dc=com             |   True   |
|         ldap_filter          |                           -                           |   True   |
|          ldap_scope          |                           3                           |   True   |
|        ldap_search_dn        |                           -                           |   True   |
|         ldap_timeout         |                           5                           |   True   |
|           ldap_uid           |                          uid                          |   True   |
|           ldap_url           |               ldaps://ldap.mydomain.com               |   True   |
| project_creation_restriction |                        everyone                       |   True   |
|       scan_all_policy        | {u'parameter': {u'daily_time': 0}, u'type': u'daily'} |   True   |
|      self_registration       |                          True                         |   True   |
|       token_expiration       |                           30                          |   True   |
|      verify_remote_cert      |                          True                         |   True   |
+------------------------------+-------------------------------------------------------+----------+
```

### Update user password

```
$ harbor change-password int32bit
Old password: *****
New Password: *****
Retype new Password: *****
Update password successfully.
```

### Promote a user to administrator

```
$ harbor promote int32bit
Promote user 'int32bit' as administrator successfully.
```

## Licensing

HarborClient is licensed under the MIT License, Version 2.0. See [LICENSE](./LICENSE) for the full license text.
