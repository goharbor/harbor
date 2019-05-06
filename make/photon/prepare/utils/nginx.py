import os, shutil
from fnmatch import fnmatch
from pathlib import Path

from g import config_dir, templates_dir
from utils.misc import prepare_config_dir, mark_file
from utils.jinja import render_jinja
from utils.cert import SSL_CERT_KEY_PATH, SSL_CERT_PATH

nginx_conf = os.path.join(config_dir, "nginx", "nginx.conf")
nginx_confd_dir = os.path.join(config_dir, "nginx", "conf.d")
nginx_https_conf_template = os.path.join(templates_dir, "nginx", "nginx.https.conf.jinja")
nginx_http_conf_template = os.path.join(templates_dir, "nginx", "nginx.http.conf.jinja")
nginx_template_ext_dir = os.path.join(templates_dir, 'nginx', 'ext')

cert_dir = Path(os.path.join(config_dir, "cert"))
ssl_cert_key = Path(os.path.join(cert_dir, 'server.key'))
ssl_cert_cert = Path(os.path.join(cert_dir, 'server.crt'))

CUSTOM_NGINX_LOCATION_FILE_PATTERN_HTTPS = 'harbor.https.*.conf'
CUSTOM_NGINX_LOCATION_FILE_PATTERN_HTTP = 'harbor.http.*.conf'

def prepare_nginx(config_dict):
    prepare_config_dir(nginx_confd_dir)
    render_nginx_template(config_dict)

def render_nginx_template(config_dict):
    if config_dict['protocol'] == "https":
        render_jinja(nginx_https_conf_template, nginx_conf,
            ssl_cert=SSL_CERT_PATH,
            ssl_cert_key=SSL_CERT_KEY_PATH)
        location_file_pattern = CUSTOM_NGINX_LOCATION_FILE_PATTERN_HTTPS
        cert_dir.mkdir(parents=True, exist_ok=True)
        ssl_cert_key.touch()
        ssl_cert_cert.touch()
    else:
        render_jinja(
            nginx_http_conf_template,
            nginx_conf)
        location_file_pattern = CUSTOM_NGINX_LOCATION_FILE_PATTERN_HTTP
    copy_nginx_location_configs_if_exist(nginx_template_ext_dir, nginx_confd_dir, location_file_pattern)

def add_additional_location_config(src, dst):
    """
    These conf files is used for user that wanna add additional customized locations to harbor proxy
    :params src: source of the file
    :params dst: destination file path
    """
    if not os.path.isfile(src):
        return
    print("Copying nginx configuration file {src} to {dst}".format(
        src=src, dst=dst))
    shutil.copy2(src, dst)
    mark_file(dst, mode=0o644)

def copy_nginx_location_configs_if_exist(src_config_dir, dst_config_dir, filename_pattern):
    if not os.path.exists(src_config_dir):
        return
    map(lambda filename: add_additional_location_config(
        os.path.join(src_config_dir, filename),
        os.path.join(dst_config_dir, filename)),
        [f for f in os.listdir(src_config_dir) if fnmatch(f, filename_pattern)])