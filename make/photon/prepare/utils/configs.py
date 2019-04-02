import yaml
from g import versions_file_path
from .misc import generate_random_string

def validate(conf, **kwargs):
    protocol = conf.get("protocol")
    if protocol != "https" and kwargs.get('notary_mode'):
        raise Exception(
            "Error: the protocol must be https when Harbor is deployed with Notary")
    if protocol == "https":
        if not conf.get("cert_path"):
            raise Exception("Error: The protocol is https but attribute ssl_cert is not set")
        if not conf.get("cert_key_path"):
            raise Exception("Error: The protocol is https but attribute ssl_cert_key is not set")

    # Storage validate
    valid_storage_drivers = ["filesystem", "azure", "gcs", "s3", "swift", "oss"]
    storage_provider_name = conf.get("storage_provider_name")
    if storage_provider_name not in valid_storage_drivers:
        raise Exception("Error: storage driver %s is not supported, only the following ones are supported: %s" % (
            storage_provider_name, ",".join(valid_storage_drivers)))

    storage_provider_config = conf.get("storage_provider_config") ## original is registry_storage_provider_config
    if storage_provider_name != "filesystem":
        if storage_provider_config == "":
            raise Exception(
                "Error: no provider configurations are provided for provider %s" % storage_provider_name)

    # Redis validate
    redis_host = conf.get("redis_host")
    if redis_host is None or len(redis_host) < 1:
        raise Exception(
            "Error: redis_host in harbor.cfg needs to point to an endpoint of Redis server or cluster.")

    redis_port = conf.get("redis_port")
    if redis_host is None or (redis_port < 1 or redis_port > 65535):
        raise Exception(
            "Error: redis_port in harbor.cfg needs to point to the port of Redis server or cluster.")

    redis_db_index = conf.get("redis_db_index")
    if len(redis_db_index.split(",")) != 3:
        raise Exception(
             "Error invalid value for redis_db_index: %s. please set it as 1,2,3" % redis_db_index)

def parse_versions():
    if not versions_file_path.is_file():
        return {}
    with open('versions') as f:
        versions = yaml.load(f)
    return versions

def parse_yaml_config(config_file_path):
    '''
    :param configs: config_parser object
    :returns: dict of configs
    '''

    with open(config_file_path) as f:
        configs = yaml.load(f)

    config_dict = {}
    config_dict['adminserver_url'] = "http://adminserver:8080"
    config_dict['registry_url'] = "http://registry:5000"
    config_dict['registry_controller_url'] = "http://registryctl:8080"
    config_dict['core_url'] = "http://core:8080"
    config_dict['token_service_url'] = "http://core:8080/service/token"

    config_dict['jobservice_url'] = "http://jobservice:8080"
    config_dict['clair_url'] = "http://clair:6060"
    config_dict['notary_url'] = "http://notary-server:4443"
    config_dict['chart_repository_url'] = "http://chartmuseum:9999"

    if configs.get("reload_config"):
        config_dict['reload_config'] = configs.get("reload_config")
    else:
        config_dict['reload_config'] = "false"

    config_dict['hostname'] = configs.get("hostname")
    config_dict['protocol'] = configs.get("ui_url_protocol")
    config_dict['public_url'] = config_dict['protocol'] + "://" + config_dict['hostname']

    # Data path volume
    config_dict['data_volume'] = configs.get("data_volume")

    # Email related configs
    config_dict['email_identity'] = configs.get("email_identity")
    config_dict['email_host'] = configs.get("email_server")
    config_dict['email_port'] = configs.get("email_server_port")
    config_dict['email_usr'] = configs.get("email_username")
    config_dict['email_pwd'] = configs.get("email_password")
    config_dict['email_from'] = configs.get("email_from")
    config_dict['email_ssl'] = configs.get("email_ssl")
    config_dict['email_insecure'] = configs.get("email_insecure")
    config_dict['harbor_admin_password'] = configs.get("harbor_admin_password")
    config_dict['auth_mode'] = configs.get("auth_mode")
    config_dict['ldap_url'] = configs.get("ldap_url")

    # LDAP related configs
    # this two options are either both set or unset
    if configs.get("ldap_searchdn"):
        config_dict['ldap_searchdn'] = configs["ldap_searchdn"]
        config_dict['ldap_search_pwd'] = configs["ldap_search_pwd"]
    else:
        config_dict['ldap_searchdn'] = ""
        config_dict['ldap_search_pwd'] = ""
    config_dict['ldap_basedn'] = configs.get("ldap_basedn")
    # ldap_filter is null by default
    if configs.get("ldap_filter"):
        config_dict['ldap_filter'] = configs["ldap_filter"]
    else:
        config_dict['ldap_filter'] = ""
    config_dict['ldap_uid'] = configs.get("ldap_uid")
    config_dict['ldap_scope'] = configs.get("ldap_scope")
    config_dict['ldap_timeout'] = configs.get("ldap_timeout")
    config_dict['ldap_verify_cert'] = configs.get("ldap_verify_cert")
    config_dict['ldap_group_basedn'] = configs.get("ldap_group_basedn")
    config_dict['ldap_group_filter'] = configs.get("ldap_group_filter")
    config_dict['ldap_group_gid'] = configs.get("ldap_group_gid")
    config_dict['ldap_group_scope'] = configs.get("ldap_group_scope")
    # Admin dn
    config_dict['ldap_group_admin_dn'] = configs.get("ldap_group_admin_dn") or ''

    # DB configs
    db_configs = configs.get('database')
    config_dict['db_host'] = db_configs.get("host")
    config_dict['db_port'] = db_configs.get("port")
    config_dict['db_user'] = db_configs.get("username")
    config_dict['db_password'] = db_configs.get("password")

    config_dict['self_registration'] = configs.get("self_registration")
    config_dict['project_creation_restriction'] = configs.get("project_creation_restriction")

    # secure configs
    if config_dict['protocol'] == "https":
        config_dict['cert_path'] = configs.get("ssl_cert")
        config_dict['cert_key_path'] = configs.get("ssl_cert_key")
    config_dict['customize_crt'] = configs.get("customize_crt")
    config_dict['max_job_workers'] = configs.get("max_job_workers")
    config_dict['token_expiration'] = configs.get("token_expiration")

    config_dict['secretkey_path'] = configs["secretkey_path"]
     # Admiral configs
    if configs.get("admiral_url"):
        config_dict['admiral_url'] = configs["admiral_url"]
    else:
        config_dict['admiral_url'] = ""

    # Clair configs
    clair_configs = configs.get("clair") or {}
    config_dict['clair_db_password'] = clair_configs.get("db_password") or ''
    config_dict['clair_db_host'] = clair_configs.get("db_host") or ''
    config_dict['clair_db_port'] = clair_configs.get("db_port") or ''
    config_dict['clair_db_username'] = clair_configs.get("db_username") or ''
    config_dict['clair_db'] = clair_configs.get("db") or ''
    config_dict['clair_updaters_interval'] = clair_configs.get("updaters_interval") or ''
    config_dict['clair_http_proxy'] = clair_configs.get('http_proxy') or ''
    config_dict['clair_https_proxy'] = clair_configs.get('https_proxy') or ''
    config_dict['clair_no_proxy'] = clair_configs.get('no_proxy') or ''

    # UAA configs
    config_dict['uaa_endpoint'] = configs.get("uaa_endpoint")
    config_dict['uaa_clientid'] = configs.get("uaa_clientid")
    config_dict['uaa_clientsecret'] = configs.get("uaa_clientsecret")
    config_dict['uaa_verify_cert'] = configs.get("uaa_verify_cert")
    config_dict['uaa_ca_cert'] = configs.get("uaa_ca_cert")

    # Log configs
    log_configs = configs.get('log') or {}
    config_dict['log_location'] = log_configs.get("location")
    config_dict['log_rotate_count'] = log_configs.get("rotate_count")
    config_dict['log_rotate_size'] = log_configs.get("rotate_size")

    # Redis configs
    redis_configs = configs.get("redis")
    if redis_configs:
        config_dict['redis_host'] = redis_configs.get("host") or ''
        config_dict['redis_port'] = redis_configs.get("port") or ''
        config_dict['redis_password'] = redis_configs.get("password") or ''
        config_dict['redis_db_index'] = redis_configs.get("db_index") or ''
        db_indexs = config_dict['redis_db_index'].split(',')
        config_dict['redis_db_index_reg'] = db_indexs[0]
        config_dict['redis_db_index_js'] = db_indexs[1]
        config_dict['redis_db_index_chart'] = db_indexs[2]
    else:
        config_dict['redis_host'] = ''
        config_dict['redis_port'] = ''
        config_dict['redis_password'] = ''
        config_dict['redis_db_index'] = ''
        config_dict['redis_db_index_reg'] = ''
        config_dict['redis_db_index_js'] = ''
        config_dict['redis_db_index_chart'] = ''

    # redis://[arbitrary_username:password@]ipaddress:port/database_index
    if config_dict.get('redis_password'):
        config_dict['redis_url_js'] = "redis://anonymous:%s@%s:%s/%s" % (config_dict['redis_password'], config_dict['redis_host'], config_dict['redis_port'], config_dict['redis_db_index_js'])
        config_dict['redis_url_reg'] = "redis://anonymous:%s@%s:%s/%s" % (config_dict['redis_password'], config_dict['redis_host'], config_dict['redis_port'], config_dict['redis_db_index_reg'])
    else:
        config_dict['redis_url_js'] = "redis://%s:%s/%s" % (config_dict['redis_host'], config_dict['redis_port'], config_dict['redis_db_index_js'])
        config_dict['redis_url_reg'] = "redis://%s:%s/%s" % (config_dict['redis_host'], config_dict['redis_port'], config_dict['redis_db_index_reg'])

    if configs.get("skip_reload_env_pattern"):
        config_dict['skip_reload_env_pattern'] = configs["skip_reload_env_pattern"]
    else:
        config_dict['skip_reload_env_pattern'] = "$^"

    # Registry storage configs
    storage_config = configs.get('storage')
    if storage_config:
        config_dict['storage_provider_name'] = storage_config.get("registry_storage_provider_name") or ''
        config_dict['storage_provider_config'] = storage_config.get("registry_storage_provider_config") or ''
        # yaml requires 1 or more spaces between the key and value
        config_dict['storage_provider_config'] = config_dict['storage_provider_config'].replace(":", ": ", 1)
        config_dict['registry_custom_ca_bundle_path'] = storage_config.get("registry_custom_ca_bundle") or ''
    else:
        config_dict['storage_provider_name'] = ''
        config_dict['storage_provider_config'] = ''
        config_dict['registry_custom_ca_bundle_path'] = ''

    # auto generate secret string
    config_dict['core_secret'] = generate_random_string(16)
    config_dict['jobservice_secret'] = generate_random_string(16)

    return config_dict