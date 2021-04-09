# coding: utf-8

from library.base import Base

class Health(Base, object):
    def __init__(self):
        super(Health,self).__init__(api_type = "health")
    def testHealthCheck(self):
        status, code, _ = self._get_client(**kwargs).get_health_with_http_info()
        self.assertEqual(code, 200)
        self.assertEqual("healthy", status.status)
