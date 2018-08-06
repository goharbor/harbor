import time
import os
import sys
sys.path.append(os.environ["SWAGGER_CLIENT_PATH"])
from swagger_client.rest import ApiException
import swagger_client.models
from pprint import pprint

def GetProductApi(username, password, harbor_server= os.environ["HARBOR_HOST"]):
    
    cfg = swagger_client.Configuration()
    cfg.host = "https://"+harbor_server+"/api"
    cfg.username = username
    cfg.password = password
    cfg.verify_ssl = False
    cfg.debug = True
    api_client = swagger_client.ApiClient(cfg)
    api_instance = swagger_client.ProductsApi(api_client)
    return api_instance
