import os, shutil, pathlib
from g import templates_dir, config_dir, root_crt_path, secret_key_dir,DEFAULT_UID, DEFAULT_GID
from .cert import openssl_installed, create_cert, create_root_cert, get_alias
from .jinja import render_jinja
from .misc import mark_file, prepare_config_dir

notary_template_dir = os.path.join(templates_dir, "notary")
notary_signer_pg_template = os.path.join(notary_template_dir, "signer-config.postgres.json.jinja")
notary_server_pg_template = os.path.join(notary_template_dir, "server-config.postgres.json.jinja")
notary_server_nginx_config_template = os.path.join(templates_dir, "nginx", "notary.server.conf.jinja")
notary_signer_env_template = os.path.join(notary_template_dir, "signer_env.jinja")
notary_server_env_template = os.path.join(notary_template_dir, "server_env.jinja")

notary_config_dir = os.path.join(config_dir, 'notary')
notary_signer_pg_config = os.path.join(notary_config_dir, "signer-config.postgres.json")
notary_server_pg_config = os.path.join(notary_config_dir, "server-config.postgres.json")
notary_server_config_path = os.path.join(notary_config_dir, 'notary.server.conf')
notary_signer_env_path = os.path.join(notary_config_dir, "signer_env")
notary_server_env_path = os.path.join(notary_config_dir, "server_env")


def prepare_env_notary(nginx_config_dir):
    notary_config_dir = prepare_config_dir(config_dir, "notary")
    old_signer_cert_secret_path = pathlib.Path(os.path.join(config_dir, 'notary-signer.crt'))
    old_signer_key_secret_path = pathlib.Path(os.path.join(config_dir, 'notary-signer.key'))
    old_signer_ca_cert_secret_path = pathlib.Path(os.path.join(config_dir, 'notary-signer-ca.crt'))

    notary_secret_dir = prepare_config_dir('/secret/notary')
    signer_cert_secret_path = pathlib.Path(os.path.join(notary_secret_dir, 'notary-signer.crt'))
    signer_key_secret_path = pathlib.Path(os.path.join(notary_secret_dir, 'notary-signer.key'))
    signer_ca_cert_secret_path = pathlib.Path(os.path.join(notary_secret_dir, 'notary-signer-ca.crt'))

    # In version 1.8 the secret path changed
    # If cert, key , ca all are exist in new place don't do anything
    if not(
        signer_cert_secret_path.exists() and
        signer_key_secret_path.exists() and
        signer_ca_cert_secret_path.exists()
        ):
        # If the certs are exist in old place, move it to new place
        if old_signer_ca_cert_secret_path.exists() and old_signer_cert_secret_path.exists() and old_signer_key_secret_path.exists():
            print("Copying certs for notary signer")
            shutil.copy2(old_signer_ca_cert_secret_path, signer_ca_cert_secret_path)
            shutil.copy2(old_signer_key_secret_path, signer_key_secret_path)
            shutil.copy2(old_signer_cert_secret_path, signer_cert_secret_path)
        # If certs neither exist in new place nor in the old place, create it and move it to new place
        elif openssl_installed():
            try:
                temp_cert_dir = os.path.join('/tmp', "cert_tmp")
                if not os.path.exists(temp_cert_dir):
                    os.makedirs(temp_cert_dir)
                ca_subj = "/C=US/ST=California/L=Palo Alto/O=GoHarbor/OU=Harbor/CN=Self-signed by GoHarbor"
                cert_subj = "/C=US/ST=California/L=Palo Alto/O=GoHarbor/OU=Harbor/CN=notarysigner"
                signer_ca_cert = os.path.join(temp_cert_dir, "notary-signer-ca.crt")
                signer_ca_key = os.path.join(temp_cert_dir, "notary-signer-ca.key")
                signer_cert_path = os.path.join(temp_cert_dir, "notary-signer.crt")
                signer_key_path = os.path.join(temp_cert_dir, "notary-signer.key")
                create_root_cert(ca_subj, key_path=signer_ca_key, cert_path=signer_ca_cert)
                create_cert(cert_subj, signer_ca_key, signer_ca_cert, key_path=signer_key_path, cert_path=signer_cert_path)
                print("Copying certs for notary signer")
                shutil.copy2(signer_cert_path, signer_cert_secret_path)
                shutil.copy2(signer_key_path, signer_key_secret_path)
                shutil.copy2(signer_ca_cert, signer_ca_cert_secret_path)
            finally:
                srl_tmp = os.path.join(os.getcwd(), ".srl")
                if os.path.isfile(srl_tmp):
                    os.remove(srl_tmp)
                if os.path.isdir(temp_cert_dir):
                    shutil.rmtree(temp_cert_dir, True)
        else:
            raise(Exception("No certs for notary"))


    print("Copying nginx configuration file for notary")

    render_jinja(
        os.path.join(templates_dir, "nginx", "notary.upstream.conf.jinja"),
        os.path.join(nginx_config_dir, "notary.upstream.conf"),
        gid=DEFAULT_GID,
        uid=DEFAULT_UID)

    mark_file(os.path.join(notary_secret_dir, "notary-signer.crt"))
    mark_file(os.path.join(notary_secret_dir, "notary-signer.key"))
    mark_file(os.path.join(notary_secret_dir, "notary-signer-ca.crt"))


def prepare_notary(config_dict, nginx_config_dir, ssl_cert_path, ssl_cert_key_path):

    prepare_env_notary(nginx_config_dir)

    render_jinja(
        notary_server_nginx_config_template,
        os.path.join(nginx_config_dir, "notary.server.conf"),
        gid=DEFAULT_GID,
        uid=DEFAULT_UID,
        ssl_cert=ssl_cert_path,
        ssl_cert_key=ssl_cert_key_path)

    render_jinja(
        notary_server_pg_template,
        notary_server_pg_config,
        uid=DEFAULT_UID,
        gid=DEFAULT_GID,
        token_endpoint=config_dict['public_url'],
        **config_dict)

    render_jinja(
        notary_server_env_template,
        notary_server_env_path,
        **config_dict
    )

    default_alias = get_alias(secret_key_dir)

    render_jinja(
        notary_signer_env_template,
        notary_signer_env_path,
        alias=default_alias,
        **config_dict)

    render_jinja(
        notary_signer_pg_template,
        notary_signer_pg_config,
        uid=DEFAULT_UID,
        gid=DEFAULT_GID,
        alias=default_alias,
        **config_dict)
