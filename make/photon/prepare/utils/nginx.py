import os, shutil
from fnmatch import fnmatch
from pathlib import Path

from g import config_dir, templates_dir, host_root_dir, DEFAULT_GID, DEFAULT_UID, data_dir
from utils.misc import prepare_dir, mark_file
from utils.jinja import render_jinja
from utils.cert import SSL_CERT_KEY_PATH, SSL_CERT_PATH

host_ngx_real_cert_dir = Path(os.path.join(data_dir, 'secret', 'cert'))

nginx_conf = os.path.join(config_dir, "nginx", "nginx.conf")
nginx_confd_dir = os.path.join(config_dir, "nginx", "conf.d")
nginx_https_conf_template = os.path.join(templates_dir, "nginx", "nginx.https.conf.jinja")
nginx_http_conf_template = os.path.join(templates_dir, "nginx", "nginx.http.conf.jinja")
nginx_template_ext_dir = os.path.join(templates_dir, 'nginx', 'ext')

CUSTOM_NGINX_LOCATION_FILE_PATTERN_HTTPS = 'harbor.https.*.conf'
CUSTOM_NGINX_LOCATION_FILE_PATTERN_HTTP = 'harbor.http.*.conf'

def prepare_nginx(config_dict):
    prepare_dir(nginx_confd_dir, uid=DEFAULT_UID, gid=DEFAULT_GID)
    render_nginx_template(config_dict)


def prepare_nginx_certs(cert_key_path, cert_path):
    """
    Prepare the certs file with proper ownership
    1. Remove nginx cert files in secret dir
    2. Copy cert files on host filesystem to secret dir
    3. Change the permission to 644 and ownership to 10000:10000
    """
    host_ngx_cert_key_path = Path(os.path.join(host_root_dir, cert_key_path.lstrip('/')))
    host_ngx_cert_path = Path(os.path.join(host_root_dir, cert_path.lstrip('/')))

    if host_ngx_real_cert_dir.exists() and host_ngx_real_cert_dir.is_dir():
        shutil.rmtree(host_ngx_real_cert_dir)

    os.makedirs(host_ngx_real_cert_dir, mode=0o755)
    real_key_path = os.path.join(host_ngx_real_cert_dir, 'server.key')
    real_crt_path = os.path.join(host_ngx_real_cert_dir, 'server.crt')
    shutil.copy2(host_ngx_cert_key_path, real_key_path)
    shutil.copy2(host_ngx_cert_path, real_crt_path)

    os.chown(host_ngx_real_cert_dir, uid=DEFAULT_UID, gid=DEFAULT_GID)
    mark_file(real_key_path, uid=DEFAULT_UID, gid=DEFAULT_GID)
    mark_file(real_crt_path, uid=DEFAULT_UID, gid=DEFAULT_GID)


def render_nginx_template(config_dict):
    """
    1. render nginx config file through protocol
    2. copy additional configs to cert.d dir
    """
    if config_dict['protocol'] == 'https':
        prepare_nginx_certs(config_dict['cert_key_path'], config_dict['cert_path'])
        render_jinja(
            nginx_https_conf_template,
            nginx_conf,
            uid=DEFAULT_UID,
            gid=DEFAULT_GID,
            https_redirect='$host' + ('https_port' in config_dict and (":" + str(config_dict['https_port'])) or ""),
            ssl_cert=SSL_CERT_PATH,
            ssl_cert_key=SSL_CERT_KEY_PATH)
        location_file_pattern = CUSTOM_NGINX_LOCATION_FILE_PATTERN_HTTPS

    else:
        render_jinja(
            nginx_http_conf_template,
            nginx_conf,
            uid=DEFAULT_UID,
            gid=DEFAULT_GID)
        location_file_pattern = CUSTOM_NGINX_LOCATION_FILE_PATTERN_HTTP
    copy_nginx_location_configs_if_exist(nginx_template_ext_dir, nginx_confd_dir, location_file_pattern)


def copy_nginx_location_configs_if_exist(src_config_dir, dst_config_dir, filename_pattern):
    if not os.path.exists(src_config_dir):
        return

    def add_additional_location_config(src, dst):
        """
        These conf files is used for user that wanna add additional customized locations to harbor proxy
        :params src: source of the file
        :params dst: destination file path
        """
        if not os.path.isfile(src):
            return
        print("Copying nginx configuration file {src} to {dst}".format(src=src, dst=dst))
        shutil.copy2(src, dst)
        mark_file(dst, mode=0o644)

    map(lambda filename: add_additional_location_config(
        os.path.join(src_config_dir, filename),
        os.path.join(dst_config_dir, filename)),
        [f for f in os.listdir(src_config_dir) if fnmatch(f, filename_pattern)])