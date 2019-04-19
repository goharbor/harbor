import os, shutil

from g import templates_dir, config_dir
from .jinja import render_jinja

chartm_temp_dir = os.path.join(templates_dir, "chartserver")
chartm_env_temp = os.path.join(chartm_temp_dir, "env.jinja")

chartm_config_dir = os.path.join(config_dir, "chartserver")
chartm_env = os.path.join(config_dir, "chartserver", "env")

def prepare_chartmuseum(config_dict):

    core_secret = config_dict['core_secret']
    registry_custom_ca_bundle_path = config_dict['registry_custom_ca_bundle_path']
    redis_host = config_dict['redis_host']
    redis_port = config_dict['redis_port']
    redis_password = config_dict['redis_password']
    redis_db_index_chart = config_dict['redis_db_index_chart']
    storage_provider_config = config_dict['storage_provider_config']
    storage_provider_name = config_dict['storage_provider_name']

    if not os.path.isdir(chartm_config_dir):
        print ("Create config folder: %s" % chartm_config_dir)
        os.makedirs(chartm_config_dir)

    # process redis info
    cache_store = "redis"
    cache_redis_password = redis_password
    cache_redis_addr = "{}:{}".format(redis_host, redis_port)
    cache_redis_db_index = redis_db_index_chart


    # process storage info
    #default using local file system
    storage_driver = "local"
    # storage provider configurations
    # please be aware that, we do not check the validations of the values for the specified keys
    # convert the configs to config map
    storgae_provider_confg_map = storage_provider_config
    storage_provider_config_options = []

    if storage_provider_name == "s3":
        # aws s3 storage
        storage_driver = "amazon"
        storage_provider_config_options.append("STORAGE_AMAZON_BUCKET=%s" % storgae_provider_confg_map.get("bucket", ""))
        storage_provider_config_options.append("STORAGE_AMAZON_PREFIX=%s" % storgae_provider_confg_map.get("rootdirectory", ""))
        storage_provider_config_options.append("STORAGE_AMAZON_REGION=%s" % storgae_provider_confg_map.get("region", ""))
        storage_provider_config_options.append("STORAGE_AMAZON_ENDPOINT=%s" % storgae_provider_confg_map.get("regionendpoint", ""))
        storage_provider_config_options.append("AWS_ACCESS_KEY_ID=%s" % storgae_provider_confg_map.get("accesskey", ""))
        storage_provider_config_options.append("AWS_SECRET_ACCESS_KEY=%s" % storgae_provider_confg_map.get("secretkey", ""))
    elif storage_provider_name == "gcs":
        # google cloud storage
        storage_driver = "google"
        storage_provider_config_options.append("STORAGE_GOOGLE_BUCKET=%s" % storgae_provider_confg_map.get("bucket", ""))
        storage_provider_config_options.append("STORAGE_GOOGLE_PREFIX=%s" % storgae_provider_confg_map.get("rootdirectory", ""))

        keyFileOnHost = storgae_provider_confg_map.get("keyfile", "")
        if os.path.isfile(keyFileOnHost):
            shutil.copyfile(keyFileOnHost, os.path.join(chartm_config_dir, "gcs.key"))
            targetKeyFile = "/etc/chartserver/gcs.key"
            storage_provider_config_options.append("GOOGLE_APPLICATION_CREDENTIALS=%s" % targetKeyFile)
    elif storage_provider_name == "azure":
        # azure storage
        storage_driver = "microsoft"
        storage_provider_config_options.append("STORAGE_MICROSOFT_CONTAINER=%s" % storgae_provider_confg_map.get("container", ""))
        storage_provider_config_options.append("AZURE_STORAGE_ACCOUNT=%s" % storgae_provider_confg_map.get("accountname", ""))
        storage_provider_config_options.append("AZURE_STORAGE_ACCESS_KEY=%s" % storgae_provider_confg_map.get("accountkey", ""))
        storage_provider_config_options.append("STORAGE_MICROSOFT_PREFIX=/azure/harbor/charts")
    elif storage_provider_name == "swift":
        # open stack swift
        storage_driver = "openstack"
        storage_provider_config_options.append("STORAGE_OPENSTACK_CONTAINER=%s" % storgae_provider_confg_map.get("container", ""))
        storage_provider_config_options.append("STORAGE_OPENSTACK_PREFIX=%s" % storgae_provider_confg_map.get("rootdirectory", ""))
        storage_provider_config_options.append("STORAGE_OPENSTACK_REGION=%s" % storgae_provider_confg_map.get("region", ""))
        storage_provider_config_options.append("OS_AUTH_URL=%s" % storgae_provider_confg_map.get("authurl", ""))
        storage_provider_config_options.append("OS_USERNAME=%s" % storgae_provider_confg_map.get("username", ""))
        storage_provider_config_options.append("OS_PASSWORD=%s" % storgae_provider_confg_map.get("password", ""))
        storage_provider_config_options.append("OS_PROJECT_ID=%s" % storgae_provider_confg_map.get("tenantid", ""))
        storage_provider_config_options.append("OS_PROJECT_NAME=%s" % storgae_provider_confg_map.get("tenant", ""))
        storage_provider_config_options.append("OS_DOMAIN_ID=%s" % storgae_provider_confg_map.get("domainid", ""))
        storage_provider_config_options.append("OS_DOMAIN_NAME=%s" % storgae_provider_confg_map.get("domain", ""))
    elif storage_provider_name == "oss":
        # aliyun OSS
        storage_driver = "alibaba"
        storage_provider_config_options.append("STORAGE_ALIBABA_BUCKET=%s" % storgae_provider_confg_map.get("bucket", ""))
        storage_provider_config_options.append("STORAGE_ALIBABA_PREFIX=%s" % storgae_provider_confg_map.get("rootdirectory", ""))
        storage_provider_config_options.append("STORAGE_ALIBABA_ENDPOINT=%s" % storgae_provider_confg_map.get("endpoint", ""))
        storage_provider_config_options.append("ALIBABA_CLOUD_ACCESS_KEY_ID=%s" % storgae_provider_confg_map.get("accesskeyid", ""))
        storage_provider_config_options.append("ALIBABA_CLOUD_ACCESS_KEY_SECRET=%s" % storgae_provider_confg_map.get("accesskeysecret", ""))
    else:
        # use local file system
        storage_provider_config_options.append("STORAGE_LOCAL_ROOTDIR=/chart_storage")

    # generate storage provider configuration
    all_storage_provider_configs = ('\n').join(storage_provider_config_options)

    render_jinja(
    chartm_env_temp,
    chartm_env,
    cache_store=cache_store,
    cache_redis_addr=cache_redis_addr,
    cache_redis_password=cache_redis_password,
    cache_redis_db_index=cache_redis_db_index,
    core_secret=core_secret,
    storage_driver=storage_driver,
    all_storage_driver_configs=all_storage_provider_configs)