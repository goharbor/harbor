import json

from jinja2 import Environment, FileSystemLoader
from .misc import mark_file

jinja_env = Environment(loader=FileSystemLoader('/'), trim_blocks=True, lstrip_blocks=True)

def to_json(value):
    return json.dumps(value)

jinja_env.filters['to_json'] = to_json


def render_jinja(src, dest,mode=0o640, uid=0, gid=0, **kw):
    t = jinja_env.get_template(src)
    with open(dest, 'w') as f:
        f.write(t.render(**kw))
    mark_file(dest, mode, uid, gid)
    print("Generated configuration file: %s" % dest)