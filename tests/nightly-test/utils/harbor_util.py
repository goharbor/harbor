from urllib2 import urlopen
import ssl
import time
import os
try:
    import json
except ImportError:
    import simplejson as json

import requests
from requests.packages.urllib3.exceptions import InsecureRequestWarning
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

def request(harbor_endpoint, url, method, user, pwd, **kwargs):
    url = "https://" + harbor_endpoint + "/api" + url
    kwargs.setdefault('headers', kwargs.get('headers', {}))
    kwargs['headers']['Accept'] = 'application/json'
    if 'body' in kwargs:
        kwargs['headers']['Content-Type'] = 'application/json'
        kwargs['data'] = json.dumps(kwargs['body'])
        del kwargs['body']

    resp = requests.request(method, url, verify=False, auth=(user, pwd), **kwargs)
    if resp.status_code >= 400:
        raise Exception("Error: %s" % resp.text)
    try:
        body = json.loads(resp.text)
    except ValueError:
        body = resp.text
    return body

# wait for 10 minutes as OVA needs about 7 minutes to startup harbor.
def wait_for_harbor_ready(harbor_endpoint, timeout=600):
    ctx = ssl.create_default_context()
    ctx.check_hostname = False
    ctx.verify_mode = ssl.CERT_NONE
    interval = 10
    while True:
        try:
            if timeout <= 0:
                return False
            code = urlopen(harbor_endpoint, context=ctx).code
            if code == 200:
                return True
        except Exception, e:
            timeout -= interval
            time.sleep(interval)
            continue
        timeout -= interval
        time.sleep(interval)

def get_harbor_version(harbor_endpoint, harbor_user, harbor_pwd):
   return request(harbor_endpoint, '/systeminfo', 'get', harbor_user, harbor_pwd)['harbor_version']
