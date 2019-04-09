import shutil, os

from g import config_dir, templates_dir
from utils.misc import prepare_config_dir, generate_random_string
from utils.jinja import render_jinja

core_config_dir = os.path.join(config_dir, "core", "certificates")
core_env_template_path = os.path.join(templates_dir, "core", "env.jinja")
core_conf_env = os.path.join(config_dir, "core", "env")
core_conf_template_path = os.path.join(templates_dir, "core", "app.conf.jinja")
core_conf = os.path.join(config_dir, "core", "app.conf")

core_config_env_template = os.path.join(templates_dir, "core", "config_env.jinja")
core_config_env = os.path.join(config_dir, "core", "config_env")

def prepare_core(config_dict, with_notary, with_clair, with_chartmuseum):
    prepare_core_config_dir()
    # Render Core
    # set cache for chart repo server
    # default set 'memory' mode, if redis is configured then set to 'redis'
    if len(config_dict['redis_host']) > 0:
        chart_cache_driver = "redis"
    else:
        chart_cache_driver = "memory"

    render_config_env(config_dict, with_notary, with_clair, with_chartmuseum)

    render_jinja(
        core_env_template_path,
        core_conf_env,
        chart_cache_driver=chart_cache_driver,
        **config_dict)

    # Copy Core app.conf
    copy_core_config(core_conf_template_path, core_conf)

def prepare_core_config_dir():
    prepare_config_dir(core_config_dir)

def copy_core_config(core_templates_path, core_config_path):
    shutil.copyfile(core_templates_path, core_config_path)
    print("Generated configuration file: %s" % core_config_path)

def render_config_env(config_dict, with_notary, with_clair, with_chartmuseum):
    # Use reload_key to avoid reload config after restart harbor

    render_jinja(
        core_config_env_template,
        core_config_env,
        with_notary=with_notary,
        with_clair=with_clair,
        with_chartmuseum=with_chartmuseum,
        **config_dict
        )