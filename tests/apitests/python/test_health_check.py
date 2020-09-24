# coding: utf-8

from __future__ import absolute_import

import unittest
import testutils

class TestHealthCheck(unittest.TestCase):
    def testHealthCheck(self):
        client = testutils.GetProductApi("admin", "Harbor12345")
        status, code, _ = client.health_get_with_http_info()
        self.assertEqual(code, 200)
        self.assertEqual("healthy", status.status)

if __name__ == '__main__':
    unittest.main()
