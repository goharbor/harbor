import os
from jinja2 import Environment, FileSystemLoader, StrictUndefined
from migration import utils

revision = '1.9.0'
down_revision = None

def migrate(input_cfg, output_cfg):
    config_dict = utils.read_conf(input_cfg)

    this_dir = os.path.dirname(__file__)
    tpl = Environment(
        loader=FileSystemLoader(this_dir),
        undefined=StrictUndefined,
        trim_blocks=True,
        lstrip_blocks=True
        ).get_template('harbor.yml.jinja')

    with open(output_cfg, 'w') as f:
        f.write(tpl.render(**config_dict))