import os
from jinja2 import Environment, FileSystemLoader, StrictUndefined
from utils.migration import read_conf

revision = '2.1.0'
down_revisions = ['2.0.0']

def migrate(input_cfg, output_cfg):
    def max_comm_need_update(db_conf):
        # update value if database configured and max_open_conns value between 100 and 1000.
        if db_conf and db_conf.get('max_open_conns'):
                return 100 <= db_conf['max_open_conns'] < 1000
        # other situations no need to update
        return False

    current_dir = os.path.dirname(__file__)
    tpl = Environment(
        loader=FileSystemLoader(current_dir),
        undefined=StrictUndefined,
        trim_blocks=True,
        lstrip_blocks=True
        ).get_template('harbor.yml.jinja')

    config_dict = read_conf(input_cfg)

    if max_comm_need_update(config_dict.get('database')):
        config_dict['database']['max_open_conns'] = 1000

    with open(output_cfg, 'w') as f:
        f.write(tpl.render(**config_dict))
