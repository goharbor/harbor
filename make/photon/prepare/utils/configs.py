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

    config_dict = {
        'adminserver_url': "http://adminserver:8080",
        'registry_url': "http://registry:5000",
        'registry_controller_url': "http://registryctl:8080",
        'core_url': "http://core:8080",
        'token_service_url': "http://core:8080/service/token",
        'jobservice_url': 'http://jobservice:8080',
        'clair_url': 'http://clair:6060',
        'notary_url': 'http://notary-server:4443',
        'chart_repository_url': 'http://chartmuseum:9999'
    }

    config_dict['hostname'] = configs.get("hostname")
    http_config = configs.get('http')
    https_config = configs.get('https')

    if https_config:
        config_dict['protocol'] = 'https'
        config_dict['https_port'] = https_config.get('port', 443)
        config_dict['cert_path'] = https_config.get("certificate")
        config_dict['cert_key_path'] = https_config.get("private_key")
    else:
        config_dict['protocol'] = 'http'
        config_dict['http_port'] = http_config.get('port', 80)

    if configs.get('external_url'):
        config_dict['public_url'] = configs['external_url']
    else:
        config_dict['public_url'] = '{protocol}://{hostname}'.format(**config_dict)


    # DB configs
    db_configs = configs.get('database')
    if db_configs:
        config_dict['db_host'] = 'postgresql'
        config_dict['db_port'] = 5432
        config_dict['db_user'] = 'postgres'
        config_dict['db_password'] = db_configs.get("password") or 'root123'
        config_dict['ssl_mode'] = 'disable'


    # Data path volume
    config_dict['data_volume'] = configs.get('data_volume')

    # Initial Admin Password
    config_dict['harbor_admin_password'] = configs.get("harbor_admin_password")

    # Registry storage configs
    storage_config = configs.get('storage_service') or {}
    if configs.get('filesystem'):
        print('handle filesystem')
    elif configs.get('azure'):
        print('handle azure')
    elif configs.get('gcs'):
        print('handle gcs')
    elif configs.get('s3'):
        print('handle s3')
    elif configs.get('swift'):
        print('handle swift')
    elif configs.get('oss'):
        print('handle oss')
    else:
        config_dict['storage_provider_name'] = 'filesystem'
        config_dict['storage_provider_config'] = ''
        config_dict['registry_custom_ca_bundle_path'] = storage_config.get("ca_bundle") or ''


    # config_dict['storage_provider_name'] = storage_config.get("registry_storage_provider_name") or ''
    # config_dict['storage_provider_config'] = storage_config.get("registry_storage_provider_config") or ''
    # # yaml requires 1 or more spaces between the key and value
    # config_dict['storage_provider_config'] = config_dict['storage_provider_config'].replace(":", ": ", 1)
    # config_dict['registry_custom_ca_bundle_path'] = storage_config.get("registry_custom_ca_bundle") or ''


    # Clair configs
    clair_configs = configs.get("clair") or {}
    config_dict['clair_db'] = 'postgres'
    config_dict['clair_updaters_interval'] = clair_configs.get("updaters_interval") or 12
    config_dict['clair_http_proxy'] = clair_configs.get('http_proxy') or ''
    config_dict['clair_https_proxy'] = clair_configs.get('https_proxy') or ''
    config_dict['clair_no_proxy'] = clair_configs.get('no_proxy') or ''


    # jobservice config
    js_config = configs.get('jobservice', {})
    config_dict['max_job_workers'] = js_config.get("max_job_workers", 10)
    config_dict['jobservice_secret'] = generate_random_string(16)


    # Log configs
    log_configs = configs.get('log') or {}
    config_dict['log_location'] = log_configs.get("location")
    config_dict['log_rotate_count'] = log_configs.get("rotate_count")
    config_dict['log_rotate_size'] = log_configs.get("rotate_size")
    config_dict['log_level'] = log_configs.get('level')


    # external DB, if external_db enabled, it will cover the database config
    external_db_configs = configs.get('external_database')
    if external_db_configs:
        config_dict['db_password'] = external_db_configs.get('password') or 'root123'
        if external_db_configs.get('host'):
            config_dict['db_host'] = external_db_configs['host']
        if external_db_configs.get('port'):
            config_dict['db_port'] = external_db_configs['port']
        if  external_db_configs.get('username'):
            config_dict['db_user'] = db_configs['username']
        if external_db_configs.get('ssl_mode'):
            config_dict['db_ssl_mode'] = external_db_configs['ssl_mode']


    # external_redis configs
    redis_configs = configs.get("external_redis") or {}
    config_dict['redis_host'] = redis_configs.get("host") or 'redis'
    config_dict['redis_port'] = redis_configs.get("port") or 6379
    config_dict['redis_password'] = redis_configs.get("password") or ''
    config_dict['redis_db_index_reg'] = redis_configs.get('registry_db_index') or 1
    config_dict['redis_db_index_js'] = redis_configs.get('jobservice_db_index') or 2
    config_dict['redis_db_index_chart'] = redis_configs.get('chartmuseum_db_index') or 3

    # redis://[arbitrary_username:password@]ipaddress:port/database_index
    if config_dict.get('redis_password'):
        config_dict['redis_url_js'] = "redis://anonymous:%s@%s:%s/%s" % (config_dict['redis_password'], config_dict['redis_host'], config_dict['redis_port'], config_dict['redis_db_index_js'])
        config_dict['redis_url_reg'] = "redis://anonymous:%s@%s:%s/%s" % (config_dict['redis_password'], config_dict['redis_host'], config_dict['redis_port'], config_dict['redis_db_index_reg'])
    else:
        config_dict['redis_url_js'] = "redis://%s:%s/%s" % (config_dict['redis_host'], config_dict['redis_port'], config_dict['redis_db_index_js'])
        config_dict['redis_url_reg'] = "redis://%s:%s/%s" % (config_dict['redis_host'], config_dict['redis_port'], config_dict['redis_db_index_reg'])


    # auto generated secret string for core
    config_dict['core_secret'] = generate_random_string(16)

     # Admiral configs
    if configs.get("admiral_url"):
        config_dict['admiral_url'] = configs["admiral_url"]
    else:
        config_dict['admiral_url'] = ""

    return config_dict