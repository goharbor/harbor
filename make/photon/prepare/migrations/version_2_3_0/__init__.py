import os
from jinja2 import Environment, FileSystemLoader, StrictUndefined
from utils.migration import read_conf

revision = '2.3.0'
down_revisions = ['2.2.0']

def migrate(input_cfg, output_cfg):
    def db_conn_need_update(db_conf):
        if not db_conf:
            return False

        max_idle_conns = db_conf.get('max_idle_conns', 0)
        max_open_conns = db_conf.get('max_open_conns', 0)

        return max_idle_conns == 50 and max_open_conns == 1000

    current_dir = os.path.dirname(__file__)
    tpl = Environment(
        loader=FileSystemLoader(current_dir),
        undefined=StrictUndefined,
        trim_blocks=True,
        lstrip_blocks=True
        ).get_template('harbor.yml.jinja')

    config_dict = read_conf(input_cfg)

    if db_conn_need_update(config_dict.get('database')):
        config_dict['database']['max_idle_conns'] = 100
        config_dict['database']['max_open_conns'] = 900

    with open(output_cfg, 'w') as f:
        f.write(tpl.render(**config_dict))
