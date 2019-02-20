import os, shutil
from g import base_dir, templates_dir, config_dir, root_crt, DEFAULT_UID, DEFAULT_GID
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


def prepare_env_notary(customize_crt, nginx_config_dir):
    notary_config_dir = prepare_config_dir(config_dir, "notary")
    if (customize_crt == 'on' or customize_crt == True)  and openssl_installed():
        try:
            temp_cert_dir = os.path.join(base_dir, "cert_tmp")
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
            shutil.copy2(signer_cert_path, notary_config_dir)
            shutil.copy2(signer_key_path, notary_config_dir)
            shutil.copy2(signer_ca_cert, notary_config_dir)
        finally:
            srl_tmp = os.path.join(os.getcwd(), ".srl")
            if os.path.isfile(srl_tmp):
                os.remove(srl_tmp)
            if os.path.isdir(temp_cert_dir):
                shutil.rmtree(temp_cert_dir, True)
    else:
        print("Copying certs for notary signer")
        shutil.copy2(os.path.join(notary_template_dir, "notary-signer.crt"), notary_config_dir)
        shutil.copy2(os.path.join(notary_template_dir, "notary-signer.key"), notary_config_dir)
        shutil.copy2(os.path.join(notary_template_dir, "notary-signer-ca.crt"), notary_config_dir)

    shutil.copy2(root_crt, notary_config_dir)
    shutil.copy2(
        os.path.join(notary_template_dir, "server_env.jinja"),
        os.path.join(notary_config_dir, "server_env"))

    print("Copying nginx configuration file for notary")
    notary_nginx_upstream_template_conf = os.path.join(templates_dir, "nginx", "notary.upstream.conf.jinja")
    notary_server_nginx_config = os.path.join(nginx_config_dir, "notary.server.conf")
    shutil.copy2(notary_nginx_upstream_template_conf, notary_server_nginx_config)

    mark_file(os.path.join(notary_config_dir, "notary-signer.crt"))
    mark_file(os.path.join(notary_config_dir, "notary-signer.key"))
    mark_file(os.path.join(notary_config_dir, "notary-signer-ca.crt"))
    mark_file(os.path.join(notary_config_dir, "root.crt"))

    # print("Copying sql file for notary DB")
    # if os.path.exists(os.path.join(notary_config_dir, "postgresql-initdb.d")):
    #     shutil.rmtree(os.path.join(notary_config_dir, "postgresql-initdb.d"))
    # shutil.copytree(os.path.join(notary_temp_dir, "postgresql-initdb.d"), os.path.join(notary_config_dir, "postgresql-initdb.d"))


def prepare_notary(config_dict, nginx_config_dir, ssl_cert_path, ssl_cert_key_path):

    prepare_env_notary(config_dict['customize_crt'], nginx_config_dir)

    render_jinja(
        notary_signer_pg_template,
        notary_signer_pg_config,
        uid=DEFAULT_UID,
        gid=DEFAULT_GID
        )

    render_jinja(
        notary_server_pg_template,
        notary_server_pg_config,
        uid=DEFAULT_UID,
        gid=DEFAULT_GID,
        token_endpoint=config_dict['public_url'])

    render_jinja(
        notary_server_nginx_config_template,
        os.path.join(nginx_config_dir, "notary.server.conf"),
        ssl_cert=ssl_cert_path,
        ssl_cert_key=ssl_cert_key_path)

    default_alias = get_alias(config_dict['secretkey_path'])
    render_jinja(
        notary_signer_env_template,
        notary_signer_env_path,
        alias=default_alias)

    render_jinja(
        notary_server_env_template,
        notary_server_env_path
    )