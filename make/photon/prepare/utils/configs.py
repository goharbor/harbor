import yaml, configparser
from .misc import generate_random_string

def validate(conf, **kwargs):
    protocol = conf.get("protocol")
    if protocol != "https" and kwargs.get('notary_mode'):
        raise Exception(
            "Error: the protocol must be https when Harbor is deployed with Notary")
    if protocol == "https":
        if not conf.get("cert_path"): ## ssl_path in config
            raise Exception("Error: The protocol is https but attribute ssl_cert is not set")
        if not conf.get("cert_key_path"):
            raise Exception("Error: The protocol is https but attribute ssl_cert_key is not set")

    # Project validate
    project_creation = conf.get("project_creation_restriction")
    if project_creation != "everyone" and project_creation != "adminonly":
        raise Exception(
            "Error invalid value for project_creation_restriction: %s" % project_creation)

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


def parse_configs(config_file_path):
    '''
    :param configs: config_parser object
    :returns: dict of configs
    '''
    with open(config_file_path, 'r') as f:
        formated_config = u'[configuration]\n' + f.read()

    configs = configparser.ConfigParser()
    configs.read_string(formated_config)

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

    if configs.has_option("configuration", "reload_config"):
        config_dict['reload_config'] = configs.get("configuration", "reload_config")
    else:
        config_dict['reload_config'] = "false"
    config_dict['hostname'] = configs.get("configuration", "hostname")
    config_dict['protocol'] = configs.get("configuration", "ui_url_protocol")
    config_dict['public_url'] = config_dict['protocol'] + "://" + config_dict['hostname']

    # Data path volume
    config_dict['data_volume'] = configs.get("configuration", "data_volume")

    # Email related configs
    config_dict['email_identity'] = configs.get("configuration", "email_identity")
    config_dict['email_host'] = configs.get("configuration", "email_server")
    config_dict['email_port'] = configs.get("configuration", "email_server_port")
    config_dict['email_usr'] = configs.get("configuration", "email_username")
    config_dict['email_pwd'] = configs.get("configuration", "email_password")
    config_dict['email_from'] = configs.get("configuration", "email_from")
    config_dict['email_ssl'] = configs.get("configuration", "email_ssl")
    config_dict['email_insecure'] = configs.get("configuration", "email_insecure")
    config_dict['harbor_admin_password'] = configs.get("configuration", "harbor_admin_password")
    config_dict['auth_mode'] = configs.get("configuration", "auth_mode")
    config_dict['ldap_url'] = configs.get("configuration", "ldap_url")

    # LDAP related configs
    # this two options are either both set or unset
    if configs.has_option("configuration", "ldap_searchdn"):
        config_dict['ldap_searchdn'] = configs.get("configuration", "ldap_searchdn")
        config_dict['ldap_search_pwd'] = configs.get("configuration", "ldap_search_pwd")
    else:
        config_dict['ldap_searchdn'] = ""
        config_dict['ldap_search_pwd'] = ""
    config_dict['ldap_basedn'] = configs.get("configuration", "ldap_basedn")
    # ldap_filter is null by default
    if configs.has_option("configuration", "ldap_filter"):
        config_dict['ldap_filter'] = configs.get("configuration", "ldap_filter")
    else:
        config_dict['ldap_filter'] = ""
    config_dict['ldap_uid'] = configs.get("configuration", "ldap_uid")
    config_dict['ldap_scope'] = configs.get("configuration", "ldap_scope")
    config_dict['ldap_timeout'] = configs.get("configuration", "ldap_timeout")
    config_dict['ldap_verify_cert'] = configs.get("configuration", "ldap_verify_cert")
    config_dict['ldap_group_basedn'] = configs.get("configuration", "ldap_group_basedn")
    config_dict['ldap_group_filter'] = configs.get("configuration", "ldap_group_filter")
    config_dict['ldap_group_gid'] = configs.get("configuration", "ldap_group_gid")
    config_dict['ldap_group_scope'] = configs.get("configuration", "ldap_group_scope")

    # DB configs
    config_dict['db_password'] = configs.get("configuration", "db_password")
    config_dict['db_host'] = configs.get("configuration", "db_host")
    config_dict['db_user'] = configs.get("configuration", "db_user")
    config_dict['db_port'] = configs.get("configuration", "db_port")

    config_dict['self_registration'] = configs.get("configuration", "self_registration")
    config_dict['project_creation_restriction'] = configs.get("configuration", "project_creation_restriction")

    # secure configs
    if config_dict['protocol'] == "https":
        config_dict['cert_path'] = configs.get("configuration", "ssl_cert")
        config_dict['cert_key_path'] = configs.get("configuration", "ssl_cert_key")
    config_dict['customize_crt'] = configs.get("configuration", "customize_crt")
    config_dict['max_job_workers'] = configs.get("configuration", "max_job_workers")
    config_dict['token_expiration'] = configs.get("configuration", "token_expiration")
    config_dict['secretkey_path'] = configs.get("configuration", "secretkey_path")

     # Admiral configs
    if configs.has_option("configuration", "admiral_url"):
        config_dict['admiral_url'] = configs.get("configuration", "admiral_url")
    else:
        config_dict['admiral_url'] = ""

    # Clair configs
    config_dict['clair_db_password'] = configs.get("configuration", "clair_db_password")
    config_dict['clair_db_host'] = configs.get("configuration", "clair_db_host")
    config_dict['clair_db_port'] = configs.get("configuration", "clair_db_port")
    config_dict['clair_db_username'] = configs.get("configuration", "clair_db_username")
    config_dict['clair_db'] = configs.get("configuration", "clair_db")
    config_dict['clair_updaters_interval'] = configs.get("configuration", "clair_updaters_interval")

    # UAA configs
    config_dict['uaa_endpoint'] = configs.get("configuration", "uaa_endpoint")
    config_dict['uaa_clientid'] = configs.get("configuration", "uaa_clientid")
    config_dict['uaa_clientsecret'] = configs.get("configuration", "uaa_clientsecret")
    config_dict['uaa_verify_cert'] = configs.get("configuration", "uaa_verify_cert")
    config_dict['uaa_ca_cert'] = configs.get("configuration", "uaa_ca_cert")

    # Log configs
    config_dict['log_rotate_count'] = configs.get("configuration", "log_rotate_count")
    config_dict['log_rotate_size'] = configs.get("configuration", "log_rotate_size")

    # Redis configs
    config_dict['redis_host'] = configs.get("configuration", "redis_host")
    config_dict['redis_port'] = int(configs.get("configuration", "redis_port"))
    config_dict['redis_password'] = configs.get("configuration", "redis_password")
    config_dict['redis_db_index'] = configs.get("configuration", "redis_db_index")

    db_indexs = config_dict['redis_db_index'].split(',')
    config_dict['redis_db_index_reg'] = db_indexs[0]
    config_dict['redis_db_index_js'] = db_indexs[1]
    config_dict['redis_db_index_chart'] = db_indexs[2]

    # redis://[arbitrary_username:password@]ipaddress:port/database_index
    if len(config_dict['redis_password']) > 0:
        config_dict['redis_url_js'] = "redis://anonymous:%s@%s:%s/%s" % (config_dict['redis_password'], config_dict['redis_host'], config_dict['redis_port'], config_dict['redis_db_index_js'])
        config_dict['redis_url_reg'] = "redis://anonymous:%s@%s:%s/%s" % (config_dict['redis_password'], config_dict['redis_host'], config_dict['redis_port'], config_dict['redis_db_index_reg'])
    else:
        config_dict['redis_url_js'] = "redis://%s:%s/%s" % (config_dict['redis_host'], config_dict['redis_port'], config_dict['redis_db_index_js'])
        config_dict['redis_url_reg'] = "redis://%s:%s/%s" % (config_dict['redis_host'], config_dict['redis_port'], config_dict['redis_db_index_reg'])

    if configs.has_option("configuration", "skip_reload_env_pattern"):
        config_dict['skip_reload_env_pattern'] = configs.get("configuration", "skip_reload_env_pattern")
    else:
        config_dict['skip_reload_env_pattern'] = "$^"

    # Registry storage configs
    config_dict['storage_provider_name'] = configs.get("configuration", "registry_storage_provider_name").strip()
    config_dict['storage_provider_config'] = configs.get("configuration", "registry_storage_provider_config").strip()

    # yaml requires 1 or more spaces between the key and value
    config_dict['storage_provider_config'] = config_dict['storage_provider_config'].replace(":", ": ", 1)
    config_dict['registry_custom_ca_bundle_path'] = configs.get("configuration", "registry_custom_ca_bundle").strip()
    config_dict['core_secret'] = generate_random_string(16)
    config_dict['jobservice_secret'] = generate_random_string(16)

    # Admin dn
    config_dict['ldap_group_admin_dn'] = configs.get("configuration", "ldap_group_admin_dn") if configs.has_option("configuration", "ldap_group_admin_dn") else ""

    return config_dict


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

    # DB configs
    config_dict['db_password'] = configs.get("db_password")
    config_dict['db_host'] = configs.get("db_host")
    config_dict['db_user'] = configs.get("db_user")
    config_dict['db_port'] = configs.get("db_port")

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
    config_dict['clair_db_password'] = configs.get("clair_db_password")
    config_dict['clair_db_host'] = configs.get("clair_db_host")
    config_dict['clair_db_port'] = configs.get("clair_db_port")
    config_dict['clair_db_username'] = configs.get("clair_db_username")
    config_dict['clair_db'] = configs.get("clair_db")
    config_dict['clair_updaters_interval'] = configs.get("clair_updaters_interval")

    # UAA configs
    config_dict['uaa_endpoint'] = configs.get("uaa_endpoint")
    config_dict['uaa_clientid'] = configs.get("uaa_clientid")
    config_dict['uaa_clientsecret'] = configs.get("uaa_clientsecret")
    config_dict['uaa_verify_cert'] = configs.get("uaa_verify_cert")
    config_dict['uaa_ca_cert'] = configs.get("uaa_ca_cert")

    # Log configs
    config_dict['log_location'] = configs.get("log_location")
    config_dict['log_rotate_count'] = configs.get("log_rotate_count")
    config_dict['log_rotate_size'] = configs.get("log_rotate_size")

    # Redis configs
    config_dict['redis_host'] = configs.get("redis_host") or ''
    config_dict['redis_port'] = configs.get("redis_port") or ''
    config_dict['redis_password'] = configs.get("redis_password") or ''
    config_dict['redis_db_index'] = configs.get("redis_db_index") or ''

    db_indexs = config_dict['redis_db_index'].split(',')
    config_dict['redis_db_index_reg'] = db_indexs[0]
    config_dict['redis_db_index_js'] = db_indexs[1]
    config_dict['redis_db_index_chart'] = db_indexs[2]

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
    config_dict['storage_provider_name'] = configs.get("registry_storage_provider_name") or ''
    config_dict['storage_provider_config'] = configs.get("registry_storage_provider_config") or ''

    # yaml requires 1 or more spaces between the key and value
    config_dict['storage_provider_config'] = config_dict['storage_provider_config'].replace(":", ": ", 1)
    config_dict['registry_custom_ca_bundle_path'] = configs.get("registry_custom_ca_bundle") or ''
    config_dict['core_secret'] = generate_random_string(16)
    config_dict['jobservice_secret'] = generate_random_string(16)

    # Admin dn
    config_dict['ldap_group_admin_dn'] = configs["ldap_group_admin_dn"] if configs.get("ldap_group_admin_dn") else ""

    return config_dict