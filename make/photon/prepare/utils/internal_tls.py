
def prepare_tls(config_dict):
    config_dict['internal_tls'].prepare()
    config_dict['internal_tls'].validate()