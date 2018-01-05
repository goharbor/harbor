from urllib2 import urlopen
import ssl
import time

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