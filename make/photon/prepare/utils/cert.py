# Get or generate private key
import os, sys, subprocess, shutil
from pathlib import Path
from subprocess import DEVNULL
from functools import wraps

from .misc import mark_file
from .misc import generate_random_string

SSL_CERT_PATH = os.path.join("/etc/nginx/cert", "server.crt")
SSL_CERT_KEY_PATH = os.path.join("/etc/nginx/cert", "server.key")

input_cert = '/input/nginx/server.crt'
input_cert_key = '/input/nginx/server.key'

secret_cert_dir = '/secret/nginx'
secret_cert = '/secret/nginx/server.crt'
secret_cert_key = '/secret/nginx/server.key'

input_secret_keys_dir = '/input/keys'
secret_keys_dir = '/secret/keys'
allowed_secret_key_names = ['defaultalias', 'secretkey']

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

def copy_secret_keys():
    """
    Copy the secret keys, which used for encrypt user password, from input keys dir to secret keys dir
    """
    if os.path.isdir(input_secret_keys_dir) and os.path.isdir(secret_keys_dir):
        input_files = os.listdir(input_secret_keys_dir)
        secret_files = os.listdir(secret_keys_dir)
        files_need_copy =  [x for x in input_files if (x in allowed_secret_key_names) and (x not in secret_files) ]
        for f in files_need_copy:
            shutil.copy(f, secret_keys_dir)

def copy_ssl_cert():
    """
    Copy the ssl certs key paris, which used in nginx ssl certificate, from input dir to secret cert dir
    """
    if os.path.isfile(input_cert_key) and os.path.isfile(input_cert):
        os.makedirs(secret_cert_dir, exist_ok=True)
        shutil.copy(input_cert, secret_cert)
        shutil.copy(input_cert_key, secret_cert_key)

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
   rc = subprocess.call(["/usr/bin/openssl", "genrsa", "-out", key_path, "4096"], stdout=DEVNULL, stderr=subprocess.STDOUT)
   if rc != 0:
        return rc
   return subprocess.call(["/usr/bin/openssl", "req", "-new", "-x509", "-key", key_path,\
        "-out", cert_path, "-days", "3650", "-subj", subj], stdout=DEVNULL, stderr=subprocess.STDOUT)

@stat_decorator
def create_cert(subj, ca_key, ca_cert, key_path="./k.key", cert_path="./cert.crt"):
    cert_dir = os.path.dirname(cert_path)
    csr_path = os.path.join(cert_dir, "tmp.csr")
    rc = subprocess.call(["/usr/bin/openssl", "req", "-newkey", "rsa:4096", "-nodes","-sha256","-keyout", key_path,\
        "-out", csr_path, "-subj", subj], stdout=DEVNULL, stderr=subprocess.STDOUT)
    if rc != 0:
        return rc
    return subprocess.call(["/usr/bin/openssl", "x509", "-req", "-days", "3650", "-in", csr_path, "-CA", \
        ca_cert, "-CAkey", ca_key, "-CAcreateserial", "-out", cert_path], stdout=DEVNULL, stderr=subprocess.STDOUT)


def openssl_installed():
    shell_stat = subprocess.check_call(["/usr/bin/which", "openssl"], stdout=DEVNULL, stderr=subprocess.STDOUT)
    if shell_stat != 0:
        print("Cannot find openssl installed in this computer\nUse default SSL certificate file")
        return False
    return True


def prepare_ca(
    private_key_pem_path: Path,
    root_crt_path: Path,
    old_private_key_pem_path: Path,
    old_crt_path: Path,
    registry_custom_ca_bundle_config: Path,
    registry_custom_ca_bundle_storage_path: Path):
    if not ( private_key_pem_path.exists() and root_crt_path.exists() ):
        # From version 1.8 the cert storage path is changed
        # if old key paris not exist create new ones
        # if old key pairs exist in old place copy it to new place
        if not (old_crt_path.exists() and old_private_key_pem_path.exists()):
            private_key_pem_path.parent.mkdir(parents=True, exist_ok=True)
            root_crt_path.parent.mkdir(parents=True, exist_ok=True)

            empty_subj = "/"
            create_root_cert(empty_subj, key_path=private_key_pem_path, cert_path=root_crt_path)
            mark_file(private_key_pem_path)
            mark_file(root_crt_path)
        else:
            shutil.move(old_crt_path, root_crt_path)
            shutil.move(old_private_key_pem_path, private_key_pem_path)


    if not registry_custom_ca_bundle_storage_path.exists() and registry_custom_ca_bundle_config.exists():
        registry_custom_ca_bundle_storage_path.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(registry_custom_ca_bundle_config, registry_custom_ca_bundle_storage_path)
        mark_file(registry_custom_ca_bundle_storage_path)
        print("Copied custom ca bundle: %s" % registry_custom_ca_bundle_config)