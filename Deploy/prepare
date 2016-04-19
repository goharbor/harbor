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

hostname = rcp.get("configuration", "hostname")
ui_url = rcp.get("configuration", "ui_url_protocol") + "://" + hostname
email_server = rcp.get("configuration", "email_server")
email_server_port = rcp.get("configuration", "email_server_port")
email_username = rcp.get("configuration", "email_username")
email_password = rcp.get("configuration", "email_password")
email_from = rcp.get("configuration", "email_from")
harbor_admin_password = rcp.get("configuration", "harbor_admin_password")
auth_mode = rcp.get("configuration", "auth_mode")
ldap_url = rcp.get("configuration", "ldap_url")
ldap_basedn = rcp.get("configuration", "ldap_basedn")
db_password = rcp.get("configuration", "db_password")
self_registration = rcp.get("configuration", "self_registration")
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
    t = Template(open(src, 'r').read())
    with open(dest, 'w') as f:
        f.write(t.substitute(**kw))
    print("Generated configuration file: %s" % dest)

ui_conf_env = os.path.join(config_dir, "ui", "env")
ui_conf = os.path.join(config_dir, "ui", "app.conf")
registry_conf = os.path.join(config_dir, "registry", "config.yml")
db_conf_env = os.path.join(config_dir, "db", "env")

conf_files = [ ui_conf, ui_conf_env, registry_conf, db_conf_env ]
for f in conf_files:
    if os.path.exists(f):
        print("Clearing the configuration file: %s" % f)
        os.remove(f)

render(os.path.join(templates_dir, "ui", "env"),
        ui_conf_env,
        hostname=hostname,
        db_password=db_password,
        ui_url=ui_url,
        auth_mode=auth_mode,
        admin_pwd=harbor_admin_password,
        ldap_url=ldap_url,
        ldap_basedn=ldap_basedn,
	self_registration=self_registration)

render(os.path.join(templates_dir, "ui", "app.conf"),
        ui_conf,
        email_server=email_server,
        email_server_port=email_server_port,
        email_user_name=email_username,
        email_user_password=email_password,
        email_from=email_from,
        ui_url=ui_url)

render(os.path.join(templates_dir, "registry", "config.yml"),
        registry_conf,
        ui_url=ui_url)

render(os.path.join(templates_dir, "db", "env"),
        db_conf_env,
        db_password=db_password)

print("The configuration files are ready, please use docker-compose to start the service.")
