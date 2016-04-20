#!/usr/bin/python
# -*- coding: utf-8 -*-
from __future__ import print_function, unicode_literals # We require Python 2.6 or later
from string import Template
import os
import sys
from io import open

if sys.version_info[:3][0] == 2:
    import ConfigParser as ConfigParser
    import StringIO as StringIO

if sys.version_info[:3][0] == 3:
    import configparser as ConfigParser
    import io as StringIO

#Read configurations
conf = StringIO.StringIO()
conf.write("[configuration]\n")
conf.write(open("harbor.cfg").read())
conf.seek(0, os.SEEK_SET)
rcp = ConfigParser.RawConfigParser()
rcp.readfp(conf)

hostname = rcp.get("configuration", "hostname").strip('"')
ui_url = rcp.get("configuration", "ui_url_protocol").strip('"') + "://" + hostname
email_server = rcp.get("configuration", "email_server").strip('"')
email_server_port = rcp.get("configuration", "email_server_port").strip('"')
email_username = rcp.get("configuration", "email_username").strip('"')
email_password = rcp.get("configuration", "email_password").strip('"')
email_from = rcp.get("configuration", "email_from").strip('"')
harbor_admin_password = rcp.get("configuration", "harbor_admin_password").strip('"')
auth_mode = rcp.get("configuration", "auth_mode").strip('"')
ldap_url = rcp.get("configuration", "ldap_url").strip('"')
ldap_basedn = rcp.get("configuration", "ldap_basedn").strip('"')
db_password = rcp.get("configuration", "db_password").strip('"')
self_registration = rcp.get("configuration", "self_registration").strip('"')
customize_token = rcp.get("configuration", "customize_token").strip('"')
crt_countryname = rcp.get("configuration", "crt_countryname").strip('"')
crt_state = rcp.get("configuration", "crt_state").strip('"')
crt_name = rcp.get("configuration", "crt_name").strip('"')
crt_organizationname = rcp.get("configuration", "crt_organizationname").strip('"')
crt_organizationalunitname = rcp.get("configuration", "crt_organizationalunitname").strip('"')
########

base_dir = os.path.dirname(__file__)
config_dir = os.path.join(base_dir, "config")
templates_dir = os.path.join(base_dir, "templates")


ui_config_dir = os.path.join(config_dir,"ui")
if not os.path.exists(ui_config_dir):
    os.makedirs(os.path.join(config_dir, "ui"))

db_config_dir = os.path.join(config_dir, "db")
if not os.path.exists(db_config_dir):
    os.makedirs(os.path.join(config_dir, "db"))

def render(src, dest, **kw):
    t = Template(open(src, 'r').read().strip('echo').strip().strip('"'))
    with open(dest, 'w') as f:
        f.write(t.substitute(**kw))
    print("Generated configuration file: %s" % dest)

ui_conf_env = os.path.join(config_dir, "ui", "env")
ui_conf = os.path.join(config_dir, "ui", "app.conf")
registry_conf = os.path.join(config_dir, "registry", "config.yml")
db_conf_env = os.path.join(config_dir, "db", "env")

conf_files = [ ui_conf, ui_conf_env, registry_conf, db_conf_env ]
def rmdir(cf):
    for f in cf:
        if os.path.exists(f):
            print("Clearing the configuration file: %s" % f)
            os.remove(f)
rmdir(conf_files)

render(os.path.join(templates_dir, "ui", "env"),
        ui_conf_env,
        hostname=hostname,
        db_password=db_password,
        ui_url=ui_url,
        auth_mode=auth_mode,
        harbor_admin_password=harbor_admin_password,
        ldap_url=ldap_url,
        ldap_basedn=ldap_basedn,
	self_registration=self_registration)

render(os.path.join(templates_dir, "ui", "app.conf"),
        ui_conf,
        email_server=email_server,
        email_server_port=email_server_port,
        email_username=email_username,
        email_password=email_password,
        email_from=email_from,
        ui_url=ui_url)

render(os.path.join(templates_dir, "registry", "config.yml"),
        registry_conf,
        ui_url=ui_url)

render(os.path.join(templates_dir, "db", "env"),
        db_conf_env,
        db_password=db_password)

if customize_token == 'on':
    is_fail = 0
    private_key_gem = os.path.join(config_dir, "ui", "private_key.pem")
    root_crt = os.path.join(config_dir, "registry", "root.crt")
    token_conf_files = [ private_key_gem, root_crt ]
    rmdir(token_conf_files)
    import subprocess
    shell_status = subprocess.call(["openssl", "genrsa", "-out", private_key_gem, "4096"])
    if shell_status == 0:
        print("private_key.gem has been generated in %s/ui" % config_dir)
    else:
        print("gennerate private_key.gem fail.")
        is_fail = 1
    subj = "/C={0}/ST={1}/L={2}/O={3}/OU={4}"\
        .format(crt_countryname, crt_state, crt_name, crt_organizationname, crt_organizationalunitname)
    shell_status = subprocess.call(["openssl", "req", "-new", "-x509", "-key",\
        private_key_gem, "-out", root_crt, "-days", "3650", "-subj", subj])
    if shell_status == 0:
        print("root.crt has been generated in %s/registry" % config_dir)
    else:
        print("gennerate root.crt fail.")
        is_fail = 1
try:
    if is_fail == 1:
        print("some problem occurs.")
        sys.exit(1)
except Exception as e:
    pass
print("The configuration files are ready, please use docker-compose to start the service.")
