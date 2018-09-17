# -*- coding: utf-8 -*-

import sys
import time
import swagger_client

class Server:
    def __init__(self, endpoint, verify_ssl):
        self.endpoint = endpoint
        self.verify_ssl = verify_ssl

class Credential:
    def __init__(self, type, username, password):
        self.type = type
        self.username = username
        self.password = password

def _create_client(server, credential, debug):
    cfg = swagger_client.Configuration()
    cfg.host = server.endpoint
    cfg.verify_ssl = server.verify_ssl
    # support basic auth only for now
    cfg.username = credential.username
    cfg.password = credential.password
    cfg.debug = debug
    return swagger_client.ProductsApi(swagger_client.ApiClient(cfg))

def _random_name(prefix):
    return "%s-%d" % (prefix, int(round(time.time() * 1000)))

def _get_id_from_header(header):
    location = header["Location"]
    return location.split("/")[-1]

class Base:
    def __init__(self, 
        server = Server(endpoint="http://localhost:8080/api", verify_ssl=False),
        credential = Credential(type="basic_auth", username="admin", password="Harbor12345"),
        debug = True):
        if not isinstance(server.verify_ssl, bool):
            server.verify_ssl = server.verify_ssl == "True"
        self.server = server
        self.credential = credential
        self.debug = debug
        self.client = _create_client(server, credential, debug)

    def _get_client(self, **kwargs):
        if len(kwargs) == 0:
            return self.client
        server = self.server
        if "endpoint" in kwargs:
            server.endpoint = kwargs.get("endpoint")
        if "verify_ssl" in kwargs:
            if not isinstance(kwargs.get("verify_ssl"), bool):
                server.verify_ssl = kwargs.get("verify_ssl") == "True"
            else:
                server.verify_ssl = kwargs.get("verify_ssl")
        credential = self.credential
        if "type" in kwargs:
            credential.type = kwargs.get("type")
        if "username" in kwargs:
            credential.username = kwargs.get("username")
        if "password" in kwargs:
            credential.password = kwargs.get("password")
        return _create_client(server, credential, self.debug)