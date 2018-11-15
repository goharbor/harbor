# Get or generate private key
import os, sys, subprocess, shutil
from subprocess import DEVNULL
from functools import wraps

from .misc import mark_file
from .misc import generate_random_string

SSL_CERT_PATH = os.path.join("/etc/nginx/cert", "server.crt")
SSL_CERT_KEY_PATH = os.path.join("/etc/nginx/cert", "server.key")

def _get_secret(folder, filename, length=16):
    key_file = os.path.join(folder, filename)
    if os.path.isfile(key_file):
        with open(key_file, 'r') as f:
            key = f.read()
            print("loaded secret from file: %s" % key_file)
        mark_file(key_file)
        return key
    if not os.path.isdir(folder):
        os.makedirs(folder)
    key = generate_random_string(length)
    with open(key_file, 'w') as f:
        f.write(key)
        print("Generated and saved secret to file: %s" % key_file)
    mark_file(key_file)
    return key


def get_secret_key(path):
    secret_key = _get_secret(path, "secretkey")
    if len(secret_key) != 16:
        raise Exception("secret key's length has to be 16 chars, current length: %d" % len(secret_key))
    return secret_key


def get_alias(path):
    alias = _get_secret(path, "defaultalias", length=8)
    return alias

## decorator actions
def stat_decorator(func):
    @wraps(func)
    def check_wrapper(*args, **kw):
        stat = func(*args, **kw)
        if stat == 0:
            print("Generated certificate, key file: {key_path}, cert file: {cert_path}".format(**kw))
        else:
            print("Fail to generate key file: {key_path}, cert file: {cert_path}".format(**kw))
            sys.exit(1)
    return check_wrapper


@stat_decorator
def create_root_cert(subj, key_path="./k.key", cert_path="./cert.crt"):
   rc = subprocess.call(["openssl", "genrsa", "-out", key_path, "4096"], stdout=DEVNULL, stderr=subprocess.STDOUT)
   if rc != 0:
        return rc
   return subprocess.call(["openssl", "req", "-new", "-x509", "-key", key_path,\
        "-out", cert_path, "-days", "3650", "-subj", subj], stdout=DEVNULL, stderr=subprocess.STDOUT)

@stat_decorator
def create_cert(subj, ca_key, ca_cert, key_path="./k.key", cert_path="./cert.crt"):
    cert_dir = os.path.dirname(cert_path)
    csr_path = os.path.join(cert_dir, "tmp.csr")
    rc = subprocess.call(["openssl", "req", "-newkey", "rsa:4096", "-nodes","-sha256","-keyout", key_path,\
        "-out", csr_path, "-subj", subj], stdout=DEVNULL, stderr=subprocess.STDOUT)
    if rc != 0:
        return rc
    return subprocess.call(["openssl", "x509", "-req", "-days", "3650", "-in", csr_path, "-CA", \
        ca_cert, "-CAkey", ca_key, "-CAcreateserial", "-out", cert_path], stdout=DEVNULL, stderr=subprocess.STDOUT)


def openssl_installed():
    shell_stat = subprocess.check_call(["which", "openssl"], stdout=DEVNULL, stderr=subprocess.STDOUT)
    if shell_stat != 0:
        print("Cannot find openssl installed in this computer\nUse default SSL certificate file")
        return False
    return True


def prepare_ca(
    customize_crt,
    private_key_pem_path, private_key_pem_template,
    root_crt_path, root_cert_template_path,
    registry_custom_ca_bundle_path, registry_custom_ca_bundle_config):

    if (customize_crt == 'on' or customize_crt == True) and openssl_installed():
        empty_subj = "/"
        create_root_cert(empty_subj, key_path=private_key_pem_path, cert_path=root_crt_path)
        mark_file(private_key_pem_path)
        mark_file(root_crt_path)
    else:
        print("Copied configuration file: %s" % private_key_pem_path)
        shutil.copyfile(private_key_pem_template, private_key_pem_path)
        print("Copied configuration file: %s" % root_crt_path)
        shutil.copyfile(root_cert_template_path, root_crt_path)

    if len(registry_custom_ca_bundle_path) > 0 and os.path.isfile(registry_custom_ca_bundle_path):
        shutil.copyfile(registry_custom_ca_bundle_path, registry_custom_ca_bundle_config)
        print("Copied custom ca bundle: %s" % registry_custom_ca_bundle_config)