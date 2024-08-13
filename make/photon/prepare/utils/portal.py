from g import config_dir, DEFAULT_GID, DEFAULT_UID, templates_dir
from utils.misc import prepare_dir
from utils.jinja import render_jinja

portal_config_dir = config_dir.joinpath("portal")
portal_conf_template_path = templates_dir.joinpath("portal", "nginx.conf.jinja")
portal_conf = config_dir.joinpath("portal", "nginx.conf")
portal_conf_setting_template_path = templates_dir.joinpath("portal", "setting.json.jinja")
portal_conf_setting = config_dir.joinpath("portal", "setting.json")

def prepare_portal(config_dict):
    prepare_dir(portal_config_dir, uid=DEFAULT_UID, gid=DEFAULT_GID)

    # Render Portal nginx config
    render_jinja(
        str(portal_conf_template_path),
        portal_conf,
        internal_tls=config_dict['internal_tls'],
        ip_family=config_dict['ip_family'],
        uid=DEFAULT_UID,
        gid=DEFAULT_GID,
        strong_ssl_ciphers=config_dict['strong_ssl_ciphers']
        )

    # Render Portal setting config
    render_jinja(
        str(portal_conf_setting_template_path),
        portal_conf_setting,
        look_and_feel=config_dict['look_and_feel']
        )
