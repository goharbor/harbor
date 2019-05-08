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

    config_dict['hostname'] = configs["hostname"]

    config_dict['protocol'] = 'http'
    http_config = configs.get('http') or {}
    config_dict['http_port'] = http_config.get('port', 80)

    https_config = configs.get('https')
    if https_config:
        config_dict['protocol'] = 'https'
        config_dict['https_port'] = https_config.get('port', 443)
        config_dict['cert_path'] = https_config["certificate"]
        config_dict['cert_key_path'] = https_config["private_key"]

    config_dict['public_url'] = configs.get('external_url') or '{protocol}://{hostname}'.format(**config_dict)

    # DB configs
    db_configs = configs.get('database')
    if db_configs:
        config_dict['db_host'] = 'postgresql'
        config_dict['db_port'] = 5432
        config_dict['db_user'] = 'postgres'
        config_dict['db_password'] = db_configs.get("password") or ''
        config_dict['ssl_mode'] = 'disable'


    # Data path volume
    config_dict['data_volume'] = configs['data_volume']

    # Initial Admin Password
    config_dict['harbor_admin_password'] = configs["harbor_admin_password"]

    # Registry storage configs
    storage_config = configs.get('storage_service') or {}

    config_dict['registry_custom_ca_bundle_path'] = storage_config.get('ca_bundle') or ''

    if storage_config.get('filesystem'):
        config_dict['storage_provider_name'] = 'filesystem'
        config_dict['storage_provider_config'] = storage_config['filesystem']
    elif storage_config.get('azure'):
        config_dict['storage_provider_name'] = 'azure'
        config_dict['storage_provider_config'] = storage_config['azure']
    elif storage_config.get('gcs'):
        config_dict['storage_provider_name'] = 'gcs'
        config_dict['storage_provider_config'] = storage_config['gcs']
    elif storage_config.get('s3'):
        config_dict['storage_provider_name'] = 's3'
        config_dict['storage_provider_config'] = storage_config['s3']
    elif storage_config.get('swift'):
        config_dict['storage_provider_name'] = 'swift'
        config_dict['storage_provider_config'] = storage_config['swift']
    elif storage_config.get('oss'):
        config_dict['storage_provider_name'] = 'oss'
        config_dict['storage_provider_config'] = storage_config['oss']
    else:
        config_dict['storage_provider_name'] = 'filesystem'
        config_dict['storage_provider_config'] = {}

    if storage_config.get('redirect'):
        config_dict['storage_redirect_disabled'] = storage_config['redirect']['disabled']

    # Clair configs
    clair_configs = configs.get("clair") or {}
    config_dict['clair_db'] = 'postgres'
    config_dict['clair_updaters_interval'] = clair_configs.get("updaters_interval") or 12
    config_dict['clair_http_proxy'] = clair_configs.get('http_proxy') or ''
    config_dict['clair_https_proxy'] = clair_configs.get('https_proxy') or ''
    config_dict['clair_no_proxy'] = clair_configs.get('no_proxy') or '127.0.0.1,localhost,core,registry'

    # jobservice config
    js_config = configs.get('jobservice') or {}
    config_dict['max_job_workers'] = js_config["max_job_workers"]
    config_dict['jobservice_secret'] = generate_random_string(16)


    # Log configs
    log_configs = configs.get('log') or {}
    config_dict['log_location'] = log_configs["location"]
    config_dict['log_rotate_count'] = log_configs["rotate_count"]
    config_dict['log_rotate_size'] = log_configs["rotate_size"]
    config_dict['log_level'] = log_configs['level']


    # external DB, if external_db enabled, it will cover the database config
    external_db_configs = configs.get('external_database') or {}
    if external_db_configs:
        config_dict['db_password'] = external_db_configs.get('password') or ''
        config_dict['db_host'] = external_db_configs['host']
        config_dict['db_port'] = external_db_configs['port']
        config_dict['db_user'] = external_db_configs['username']
        if external_db_configs.get('ssl_mode'):
            config_dict['db_ssl_mode'] = external_db_configs['ssl_mode']


    # redis config
    redis_configs = configs.get("external_redis")
    if redis_configs:
        # using external_redis
        config_dict['redis_host'] = redis_configs['host']
        config_dict['redis_port'] = redis_configs['port']
        config_dict['redis_password'] = redis_configs.get("password") or ''
        config_dict['redis_db_index_reg'] = redis_configs.get('registry_db_index') or 1
        config_dict['redis_db_index_js'] = redis_configs.get('jobservice_db_index') or 2
        config_dict['redis_db_index_chart'] = redis_configs.get('chartmuseum_db_index') or 3
    else:
        ## Using local redis
        config_dict['redis_host'] = 'redis'
        config_dict['redis_port'] = 6379
        config_dict['redis_password'] = ''
        config_dict['redis_db_index_reg'] = 1
        config_dict['redis_db_index_js'] = 2
        config_dict['redis_db_index_chart'] = 3

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
    config_dict['admiral_url'] = configs.get("admiral_url") or ""

    # UAA configs
    config_dict['uaa'] = configs.get('uaa') or {}

    return config_dict