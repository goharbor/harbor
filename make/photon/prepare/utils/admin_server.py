import os

from g import config_dir, templates_dir
from utils.misc import prepare_config_dir, generate_random_string
from utils.jinja import render_jinja

adminserver_config_dir = os.path.join(config_dir, 'adminserver')
adminserver_env_template = os.path.join(templates_dir, "adminserver", "env.jinja")
adminserver_conf_env = os.path.join(config_dir, "adminserver", "env")

def prepare_adminserver(config_dict, with_notary, with_clair, with_chartmuseum):
    prepare_adminserver_config_dir()
    render_adminserver(config_dict, with_notary, with_clair, with_chartmuseum)

def prepare_adminserver_config_dir():
    prepare_config_dir(adminserver_config_dir)

def render_adminserver(config_dict, with_notary, with_clair, with_chartmuseum):
    # Use reload_key to avoid reload config after restart harbor
    reload_key = generate_random_string(6) if config_dict['reload_config'] == "true" else ""

    render_jinja(
        adminserver_env_template,
        adminserver_conf_env,
        with_notary=with_notary,
        with_clair=with_clair,
        with_chartmuseum=with_chartmuseum,
        reload_key=reload_key,
        **config_dict
        )