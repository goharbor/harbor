from g import config_dir, DEFAULT_GID, DEFAULT_UID, templates_dir
from utils.misc import prepare_dir
from utils.jinja import render_jinja

portal_config_dir = config_dir.joinpath("portal")
portal_conf_template_path = templates_dir.joinpath("portal", "nginx.conf.jinja")
portal_conf = config_dir.joinpath("portal", "nginx.conf")

def prepare_portal(config_dict):
    prepare_dir(portal_config_dir, uid=DEFAULT_UID, gid=DEFAULT_GID)

    # Render Jobservice config
    render_jinja(
        str(portal_conf_template_path),
        portal_conf,
        internal_tls=config_dict['internal_tls'],
        ip_family=config_dict['ip_family'],
        uid=DEFAULT_UID,
        gid=DEFAULT_GID,
        ssl_protocols=config_dict['ssl_protocols'],
        strong_ssl_ciphers=config_dict['strong_ssl_ciphers']
        )
