import yaml
from g import versions_file_path
from .misc import generate_random_string

default_db_max_idle_conns = 2  # NOTE: https://golang.org/pkg/database/sql/#DB.SetMaxIdleConns
default_db_max_open_conns = 0  # NOTE: https://golang.org/pkg/database/sql/#DB.SetMaxOpenConns

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

    # log endpoint validate
    if ('log_ep_host' in conf) and not conf['log_ep_host']:
        raise Exception('Error: must set log endpoint host to enable external host')
    if ('log_ep_port' in conf) and not conf['log_ep_port']:
        raise Exception('Error: must set log endpoint port to enable external host')
    if ('log_ep_protocol' in conf) and (conf['log_ep_protocol'] not in ['udp', 'tcp']):
        raise Exception("Protocol in external log endpoint must be one of 'udp' or 'tcp' ")

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
        'core_local_url': "http://127.0.0.1:8080",
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

    if configs.get('external_url'):
        config_dict['public_url'] = configs.get('external_url')
    else:
        if config_dict['protocol'] == 'https':
            if config_dict['https_port'] == 443:
                config_dict['public_url'] = '{protocol}://{hostname}'.format(**config_dict)
            else:
                config_dict['public_url'] = '{protocol}://{hostname}:{https_port}'.format(**config_dict)
        else:
            if config_dict['http_port'] == 80:
                config_dict['public_url'] = '{protocol}://{hostname}'.format(**config_dict)
            else:
                config_dict['public_url'] = '{protocol}://{hostname}:{http_port}'.format(**config_dict)

    # DB configs
    db_configs = configs.get('database')
    if db_configs:
        # harbor db
        config_dict['harbor_db_host'] = 'postgresql'
        config_dict['harbor_db_port'] = 5432
        config_dict['harbor_db_name'] = 'registry'
        config_dict['harbor_db_username'] = 'postgres'
        config_dict['harbor_db_password'] = db_configs.get("password") or ''
        config_dict['harbor_db_sslmode'] = 'disable'
        config_dict['harbor_db_max_idle_conns'] = db_configs.get("max_idle_conns") or default_db_max_idle_conns
        config_dict['harbor_db_max_open_conns'] = db_configs.get("max_open_conns") or default_db_max_open_conns
        # clari db
        config_dict['clair_db_host'] = 'postgresql'
        config_dict['clair_db_port'] = 5432
        config_dict['clair_db_name'] = 'postgres'
        config_dict['clair_db_username'] = 'postgres'
        config_dict['clair_db_password'] = db_configs.get("password") or ''
        config_dict['clair_db_sslmode'] = 'disable'
        # notary signer
        config_dict['notary_signer_db_host'] = 'postgresql'
        config_dict['notary_signer_db_port'] = 5432
        config_dict['notary_signer_db_name'] = 'notarysigner'
        config_dict['notary_signer_db_username'] = 'signer'
        config_dict['notary_signer_db_password'] = 'password'
        config_dict['notary_signer_db_sslmode'] = 'disable'
        # notary server
        config_dict['notary_server_db_host'] = 'postgresql'
        config_dict['notary_server_db_port'] = 5432
        config_dict['notary_server_db_name'] = 'notaryserver'
        config_dict['notary_server_db_username'] = 'server'
        config_dict['notary_server_db_password'] = 'password'
        config_dict['notary_server_db_sslmode'] = 'disable'


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

    # Global proxy configs
    proxy_config = configs.get('proxy') or {}
    proxy_components = proxy_config.get('components') or []
    for proxy_component in proxy_components:
      config_dict[proxy_component + '_http_proxy'] = proxy_config.get('http_proxy') or ''
      config_dict[proxy_component + '_https_proxy'] = proxy_config.get('https_proxy') or ''
      config_dict[proxy_component + '_no_proxy'] = proxy_config.get('no_proxy') or '127.0.0.1,localhost,core,registry'

    # Clair configs, optional
    clair_configs = configs.get("clair") or {}
    config_dict['clair_db'] = 'postgres'
    config_dict['clair_updaters_interval'] = clair_configs.get("updaters_interval") or 12

    # Chart configs
    chart_configs = configs.get("chart") or {}
    config_dict['chart_absolute_url'] = chart_configs.get('absolute_url') or ''

    # jobservice config
    js_config = configs.get('jobservice') or {}
    config_dict['max_job_workers'] = js_config["max_job_workers"]
    config_dict['jobservice_secret'] = generate_random_string(16)

    # notification config
    notification_config = configs.get('notification') or {}
    config_dict['notification_webhook_job_max_retry'] = notification_config["webhook_job_max_retry"]

    # Log configs
    allowed_levels = ['debug', 'info', 'warning', 'error', 'fatal']
    log_configs = configs.get('log') or {}

    log_level = log_configs['level']
    if log_level not in allowed_levels:
        raise Exception('log level must be one of debug, info, warning, error, fatal')
    config_dict['log_level'] = log_level.lower()

    # parse local log related configs
    local_logs = log_configs.get('local') or {}
    if local_logs:
        config_dict['log_location'] = local_logs.get('location') or '/var/log/harbor'
        config_dict['log_rotate_count'] = local_logs.get('rotate_count') or 50
        config_dict['log_rotate_size'] = local_logs.get('rotate_size') or '200M'

    # parse external log endpoint related configs
    if log_configs.get('external_endpoint'):
        config_dict['log_external'] = True
        config_dict['log_ep_protocol'] = log_configs['external_endpoint']['protocol']
        config_dict['log_ep_host'] = log_configs['external_endpoint']['host']
        config_dict['log_ep_port'] = log_configs['external_endpoint']['port']
    else:
        config_dict['log_external'] = False

    # external DB, optional, if external_db enabled, it will cover the database config
    external_db_configs = configs.get('external_database') or {}
    if external_db_configs:
        # harbor db
        config_dict['harbor_db_host'] = external_db_configs['harbor']['host']
        config_dict['harbor_db_port'] = external_db_configs['harbor']['port']
        config_dict['harbor_db_name'] = external_db_configs['harbor']['db_name']
        config_dict['harbor_db_username'] = external_db_configs['harbor']['username']
        config_dict['harbor_db_password'] = external_db_configs['harbor']['password']
        config_dict['harbor_db_sslmode'] = external_db_configs['harbor']['ssl_mode']
        config_dict['harbor_db_max_idle_conns'] = external_db_configs['harbor'].get("max_idle_conns") or default_db_max_idle_conns
        config_dict['harbor_db_max_open_conns'] = external_db_configs['harbor'].get("max_open_conns") or default_db_max_open_conns
        # clair db
        config_dict['clair_db_host'] = external_db_configs['clair']['host']
        config_dict['clair_db_port'] = external_db_configs['clair']['port']
        config_dict['clair_db_name'] = external_db_configs['clair']['db_name']
        config_dict['clair_db_username'] = external_db_configs['clair']['username']
        config_dict['clair_db_password'] = external_db_configs['clair']['password']
        config_dict['clair_db_sslmode'] = external_db_configs['clair']['ssl_mode']
        # notary signer
        config_dict['notary_signer_db_host'] = external_db_configs['notary_signer']['host']
        config_dict['notary_signer_db_port'] = external_db_configs['notary_signer']['port']
        config_dict['notary_signer_db_name'] = external_db_configs['notary_signer']['db_name']
        config_dict['notary_signer_db_username'] = external_db_configs['notary_signer']['username']
        config_dict['notary_signer_db_password'] = external_db_configs['notary_signer']['password']
        config_dict['notary_signer_db_sslmode'] = external_db_configs['notary_signer']['ssl_mode']
        # notary server
        config_dict['notary_server_db_host'] = external_db_configs['notary_server']['host']
        config_dict['notary_server_db_port'] = external_db_configs['notary_server']['port']
        config_dict['notary_server_db_name'] = external_db_configs['notary_server']['db_name']
        config_dict['notary_server_db_username'] = external_db_configs['notary_server']['username']
        config_dict['notary_server_db_password'] = external_db_configs['notary_server']['password']
        config_dict['notary_server_db_sslmode'] = external_db_configs['notary_server']['ssl_mode']


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
