import logging
import os
import yaml
from urllib.parse import urlencode, quote
from g import versions_file_path, host_root_dir, DEFAULT_UID, INTERNAL_NO_PROXY_DN
from models import InternalTLS, Metric, Trace, PurgeUpload, Cache, Core
from utils.misc import generate_random_string, owner_can_read, other_can_read

# NOTE: https://golang.org/pkg/database/sql/#DB.SetMaxIdleConns
default_db_max_idle_conns = 2
# NOTE: https://golang.org/pkg/database/sql/#DB.SetMaxOpenConns
default_db_max_open_conns = 0
default_https_cert_path = '/your/certificate/path'
default_https_key_path = '/your/certificate/path'

REGISTRY_USER_NAME = 'harbor_registry_user'


def validate(conf: dict, **kwargs):
    # hostname validate
    if conf.get('hostname') == '127.0.0.1':
        raise Exception("127.0.0.1 can not be the hostname")
    if conf.get('hostname') == 'reg.mydomain.com':
        raise Exception("Please specify hostname")

    # protocol validate
    protocol = conf.get("protocol")
    if protocol == "https":
        if not conf.get("cert_path") or conf["cert_path"] == default_https_cert_path:
            raise Exception("Error: The protocol is https but attribute ssl_cert is not set")
        if not conf.get("cert_key_path") or conf['cert_key_path'] == default_https_key_path:
            raise Exception("Error: The protocol is https but attribute ssl_cert_key is not set")
    if protocol == "http":
        logging.warning("WARNING: HTTP protocol is insecure. Harbor will deprecate http protocol in the future. Please make sure to upgrade to https")

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

    # original is registry_storage_provider_config
    storage_provider_config = conf.get("storage_provider_config")
    if storage_provider_name != "filesystem":
        if storage_provider_config == "":
            raise Exception(
                "Error: no provider configurations are provided for provider %s" % storage_provider_name)
    # ca_bundle validate
    if conf.get('registry_custom_ca_bundle_path'):
        registry_custom_ca_bundle_path = conf.get('registry_custom_ca_bundle_path') or ''
        if registry_custom_ca_bundle_path.startswith('/data/'):
            ca_bundle_host_path = registry_custom_ca_bundle_path
        else:
            ca_bundle_host_path = os.path.join(host_root_dir, registry_custom_ca_bundle_path.lstrip('/'))
        try:
            uid = os.stat(ca_bundle_host_path).st_uid
            st_mode = os.stat(ca_bundle_host_path).st_mode
        except Exception as e:
            logging.error(e)
            raise Exception('Can not get file info')
        err_msg = 'Cert File {} should be owned by user with uid 10000 or readable by others'.format(registry_custom_ca_bundle_path)
        if uid == DEFAULT_UID and not owner_can_read(st_mode):
            raise Exception(err_msg)
        if uid != DEFAULT_UID and not other_can_read(st_mode):
            raise Exception(err_msg)

    # TODO:
    # If user enable trust cert dir, need check if the files in this dir is readable.

    if conf.get('trace'):
        conf['trace'].validate()

    if conf.get('purge_upload'):
        conf['purge_upload'].validate()

    if conf.get('cache'):
        conf['cache'].validate()

    if conf.get('core'):
        conf['core'].validate()


def parse_versions():
    if not versions_file_path.is_file():
        return {}
    with open('versions') as f:
        versions = yaml.safe_load(f)
    return versions


def parse_yaml_config(config_file_path, with_trivy):
    '''
    :param configs: config_parser object
    :returns: dict of configs
    '''

    with open(config_file_path) as f:
        configs = yaml.safe_load(f)

    config_dict = {
        'portal_url': 'http://portal:8080',
        'registry_url': 'http://registry:5000',
        'registry_controller_url': 'http://registryctl:8080',
        'core_url': 'http://core:8080',
        'core_local_url': 'http://127.0.0.1:8080',
        'token_service_url': 'http://core:8080/service/token',
        'jobservice_url': 'http://jobservice:8080',
        'trivy_adapter_url': 'http://trivy-adapter:8080',
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
        config_dict['harbor_db_conn_max_lifetime'] = db_configs.get("conn_max_lifetime") or '5m'
        config_dict['harbor_db_conn_max_idle_time'] = db_configs.get("conn_max_idle_time") or '0'

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
        config_dict['storage_redirect_disabled'] = storage_config['redirect']['disable']

    # Global proxy configs
    proxy_config = configs.get('proxy') or {}
    proxy_components = proxy_config.get('components') or []
    no_proxy_config = proxy_config.get('no_proxy')
    all_no_proxy = INTERNAL_NO_PROXY_DN
    if no_proxy_config:
        all_no_proxy |= set(no_proxy_config.split(','))

    for proxy_component in proxy_components:
        config_dict[proxy_component + '_http_proxy'] = proxy_config.get('http_proxy') or ''
        config_dict[proxy_component + '_https_proxy'] = proxy_config.get('https_proxy') or ''
        config_dict[proxy_component + '_no_proxy'] = ','.join(all_no_proxy)

    # Trivy configs, optional
    trivy_configs = configs.get("trivy") or {}
    config_dict['trivy_github_token'] = trivy_configs.get("github_token") or ''
    config_dict['trivy_skip_update'] = trivy_configs.get("skip_update") or False
    config_dict['trivy_skip_java_db_update'] = trivy_configs.get("skip_java_db_update") or False
    config_dict['trivy_offline_scan'] = trivy_configs.get("offline_scan") or False
    config_dict['trivy_security_check'] = trivy_configs.get("security_check") or 'vuln'
    config_dict['trivy_ignore_unfixed'] = trivy_configs.get("ignore_unfixed") or False
    config_dict['trivy_insecure'] = trivy_configs.get("insecure") or False
    config_dict['trivy_timeout'] = trivy_configs.get("timeout") or '5m0s'

    # jobservice config
    js_config = configs.get('jobservice') or {}
    config_dict['max_job_workers'] = js_config["max_job_workers"]
    config_dict['max_job_duration_hours'] = js_config.get("max_job_duration_hours") or 24
    value = config_dict["max_job_duration_hours"]
    if not isinstance(value, int) or value < 24:
        config_dict["max_job_duration_hours"] = 24
    config_dict['max_job_duration_seconds'] = config_dict['max_job_duration_hours'] * 3600
    config_dict['job_loggers'] = js_config["job_loggers"]
    config_dict['logger_sweeper_duration'] = js_config["logger_sweeper_duration"]
    config_dict['jobservice_secret'] = generate_random_string(16)

    # notification config
    notification_config = configs.get('notification') or {}
    config_dict['notification_webhook_job_max_retry'] = notification_config["webhook_job_max_retry"]
    config_dict['notification_webhook_job_http_client_timeout'] = notification_config["webhook_job_http_client_timeout"]

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
        config_dict['external_database'] = True
        # harbor db
        config_dict['harbor_db_host'] = external_db_configs['harbor']['host']
        config_dict['harbor_db_port'] = external_db_configs['harbor']['port']
        config_dict['harbor_db_name'] = external_db_configs['harbor']['db_name']
        config_dict['harbor_db_username'] = external_db_configs['harbor']['username']
        config_dict['harbor_db_password'] = external_db_configs['harbor']['password']
        config_dict['harbor_db_sslmode'] = external_db_configs['harbor']['ssl_mode']
        config_dict['harbor_db_max_idle_conns'] = external_db_configs['harbor'].get("max_idle_conns") or default_db_max_idle_conns
        config_dict['harbor_db_max_open_conns'] = external_db_configs['harbor'].get("max_open_conns") or default_db_max_open_conns
        config_dict['harbor_db_conn_max_lifetime'] = external_db_configs['harbor'].get("conn_max_lifetime") or '5m'
        config_dict['harbor_db_conn_max_idle_time'] = external_db_configs['harbor'].get("conn_max_idle_time") or '0'

    else:
        config_dict['external_database'] = False

    # update redis configs
    config_dict.update(get_redis_configs(configs.get("redis", None), configs.get("external_redis", None), with_trivy))

    # auto generated secret string for core
    config_dict['core_secret'] = generate_random_string(16)

    # UAA configs
    config_dict['uaa'] = configs.get('uaa') or {}

    config_dict['registry_username'] = REGISTRY_USER_NAME
    config_dict['registry_password'] = generate_random_string(32)

    internal_tls_config = configs.get('internal_tls')
    # TLS related configs
    if internal_tls_config and internal_tls_config.get('enabled'):
        config_dict['internal_tls'] = InternalTLS(
            internal_tls_config['enabled'],
            False,
            internal_tls_config['dir'],
            configs['data_volume'],
            with_trivy=with_trivy,
            external_database=config_dict['external_database'])
    else:
        config_dict['internal_tls'] = InternalTLS()
    # the configure item apply to internal and external tls communication
    # for compatibility, user could configure the strong_ssl_ciphers either in https section or under internal_tls section,
    # but it is more reasonable to configure it in https_config
    if https_config:
        config_dict['strong_ssl_ciphers'] = https_config.get('strong_ssl_ciphers') 
    else:
        config_dict['strong_ssl_ciphers'] = False

    if internal_tls_config:
        config_dict['strong_ssl_ciphers'] = config_dict['strong_ssl_ciphers'] or internal_tls_config.get('strong_ssl_ciphers')


    # ip_family config
    config_dict['ip_family'] = configs.get('ip_family') or {'ipv4': {'enabled': True}, 'ipv6': {'enabled': False}}

    # metric configs
    metric_config = configs.get('metric')
    if metric_config:
        config_dict['metric'] = Metric(metric_config['enabled'], metric_config['port'], metric_config['path'])
    else:
        config_dict['metric'] = Metric()

    # trace configs
    trace_config = configs.get('trace')
    config_dict['trace'] = Trace(trace_config or {})

    if config_dict['internal_tls'].enabled:
        config_dict['portal_url'] = 'https://portal:8443'
        config_dict['registry_url'] = 'https://registry:5443'
        config_dict['registry_controller_url'] = 'https://registryctl:8443'
        config_dict['core_url'] = 'https://core:8443'
        config_dict['core_local_url'] = 'https://core:8443'
        config_dict['token_service_url'] = 'https://core:8443/service/token'
        config_dict['jobservice_url'] = 'https://jobservice:8443'
        config_dict['trivy_adapter_url'] = 'https://trivy-adapter:8443'

    # purge upload configs
    purge_upload_config = configs.get('upload_purging')
    config_dict['purge_upload'] = PurgeUpload(purge_upload_config or {})

    # cache configs
    cache_config = configs.get('cache')
    config_dict['cache'] = Cache(cache_config or {})

    # core configs
    core_config = configs.get('core')
    config_dict['core'] = Core(core_config or {})

    return config_dict


def get_redis_url(db, redis=None):
    """Returns redis url with format `redis://[arbitrary_username:password@]ipaddress:port/database_index?idle_timeout_seconds=30`

    >>> get_redis_url(1)
    'redis://redis:6379/1'
    >>> get_redis_url(1, {'host': 'localhost:6379', 'password': 'password'})
    'redis://anonymous:password@localhost:6379/1'
    >>> get_redis_url(1, {'host':'host1:26379,host2:26379', 'sentinel_master_set':'mymaster', 'password':'password1'})
    'redis+sentinel://anonymous:password@host1:26379,host2:26379/mymaster/1'
    >>> get_redis_url(1, {'host':'host1:26379,host2:26379', 'sentinel_master_set':'mymaster', 'password':'password1','idle_timeout_seconds':30})
    'redis+sentinel://anonymous:password@host1:26379,host2:26379/mymaster/1?idle_timeout_seconds=30'

    """
    kwargs = {
        'host': 'redis:6379',
        'password': '',
    }
    kwargs.update(redis or {})
    kwargs['scheme'] = kwargs.get('sentinel_master_set', None) and 'redis+sentinel' or 'redis'
    kwargs['db_part'] = db and ("/%s" % db) or ""
    kwargs['sentinel_part'] = kwargs.get('sentinel_master_set', None) and ("/" + kwargs['sentinel_master_set']) or ''
    kwargs['password_part'] = quote(str(kwargs.get('password', None)), safe='') and (':%s@' % quote(str(kwargs['password']), safe='')) or ''
    kwargs['username_part'] = kwargs.get('username', None) or ''

    return "{scheme}://{username_part}{password_part}{host}{sentinel_part}{db_part}".format(**kwargs) + get_redis_url_param(kwargs)


def get_redis_url_param(redis=None):
    params = {}
    if redis and 'idle_timeout_seconds' in redis:
        params['idle_timeout_seconds'] = redis['idle_timeout_seconds']
    if params:
        return "?" + urlencode(params)
    return ""


def get_redis_configs(internal_redis=None, external_redis=None, with_trivy=True):
    """Returns configs for redis

    >>> get_redis_configs()['external_redis']
    False
    >>> get_redis_configs()['redis_url_reg']
    'redis://redis:6379/1'
    >>> get_redis_configs()['redis_url_js']
    'redis://redis:6379/2'
    >>> get_redis_configs()['trivy_redis_url']
    'redis://redis:6379/5'

    >>> get_redis_configs({'host': 'localhost', 'password': ''})['redis_password']
    ''
    >>> get_redis_configs({'host': 'localhost', 'password': None})['redis_password']
    ''
    >>> get_redis_configs({'host': 'localhost', 'password': None})['redis_url_reg']
    'redis://localhost:6379/1'

    >>> get_redis_configs({'host': 'localhost', 'password': 'pass'})['external_redis']
    True
    >>> get_redis_configs({'host': 'localhost', 'password': 'pass'})['redis_password']
    'pass'
    >>> get_redis_configs({'host': 'localhost', 'password': 'pass'})['redis_url_reg']
    'redis://anonymous:pass@localhost:6379/1'
    >>> get_redis_configs({'host': 'localhost', 'password': 'pass'})['redis_url_js']
    'redis://anonymous:pass@localhost:6379/2'
    >>> get_redis_configs({'host': 'localhost', 'password': 'pass'})['trivy_redis_url']
    'redis://anonymous:pass@localhost:6379/5'

    >>> 'trivy_redis_url' not in get_redis_configs(with_trivy=False)
    True
    """

    internal_redis = internal_redis or {}
    external_redis = external_redis or {}

    configs = dict(external_redis=bool(external_redis))

    # internal redis config as the default
    redis = {
        'host': 'redis:6379',
        'password': '',
        'registry_db_index': 1,
        'jobservice_db_index': 2,
        'trivy_db_index': 5,
        'idle_timeout_seconds': 30,
    }

    if len(internal_redis) > 0:
        # overwriting existing keys by internal_redis
        redis.update({key: value for (key, value) in internal_redis.items() if value})
    else:
        # overwriting existing keys by external_redis
        redis.update({key: value for (key, value) in external_redis.items() if value})

    configs['redis_url_core'] = get_redis_url(0, redis)
    configs['redis_url_js'] = get_redis_url(redis['jobservice_db_index'], redis)
    configs['redis_url_reg'] = get_redis_url(redis['registry_db_index'], redis)

    if redis.get('harbor_db_index'):
        configs['redis_url_harbor'] = get_redis_url(redis['harbor_db_index'], redis)
    if redis.get('cache_layer_db_index'):
        configs['redis_url_cache_layer'] = get_redis_url(redis['cache_layer_db_index'], redis)

    if with_trivy:
        configs['trivy_redis_url'] = get_redis_url(redis['trivy_db_index'], redis)

    return configs
