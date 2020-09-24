# -*- coding: utf-8 -*-

import time
import base
import swagger_client
from swagger_client.rest import ApiException

class Scanner(base.Base, object):
    def __init__(self):
        super(Scanner,self).__init__(api_type = "scanner")

    def scanners_get(self, **kwargs):
        client = self._get_client(**kwargs)
        return client.scanners_get()

    def scanners_get_uuid(self, is_default = False, **kwargs):
        scanners = self.scanners_get(**kwargs)
        for scanner in scanners:
            if scanner.is_default == is_default:
                return scanner.uuid

    def scanners_registration_id_patch(self, registration_id, is_default = True, **kwargs):
        client = self._get_client(**kwargs)
        isdefault = swagger_client.IsDefault(is_default)
        client.scanners_registration_id_patch(registration_id, isdefault)

