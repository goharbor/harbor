# -*- coding: utf-8 -*-

from __future__ import absolute_import
import unittest
import json

from testutils import suppress_urllib3_warning
from library.configurations import Configurations
from library.system_info import System_info


class TestBannerMessage(unittest.TestCase):


    @suppress_urllib3_warning
    def setUp(self):
        self.configurations = Configurations()
        self.system_info = System_info()
        self.message = "This is a test message."
        self.message_type = "info"
        self.closable = True
        self.from_date = "10/27/2023"
        self.to_date = "10/31/2030"


    def testBannerMessage(self):
        """
        Test case:
            Banner Message Api
        Test step and expected result:
            1. Setup banner message;
            2. Get banner message by configurations api;
            3. Check banner message by configurations api;
            4. Get banner message by system info api;
            5. Check banner message by system info api;
            6. Reset banner message;
            7. Get banner message by configurations api;
            8. Check banner message by configurations api;
            9. Get banner message by system info api;
            10. Check banner message by system info api;
        """
        # 1. Setup banner message
        self.configurations.set_configurations_of_banner_message(message=self.message, message_type=self.message_type, closable=self.closable, from_date=self.from_date, to_date=self.to_date)

        # 2. Get banner message by configurations api
        configurations = self.configurations.get_configurations()

        # 3. Check banner message by configurations api
        config_banner_message = configurations.banner_message
        config_banner_message_value = json.loads(config_banner_message.value)
        self.assertEqual(config_banner_message.editable, True)
        self.checkBannerMseeage(config_banner_message_value)

        # 4. Get banner message by system info api
        system_info = self.system_info.get_system_info()

        # 5. Check banner message by system info api
        system_info_banner_message = json.loads(system_info.banner_message)
        self.checkBannerMseeage(system_info_banner_message)

        # 6. Reset banner message
        self.message = ""
        self.configurations.set_configurations_of_banner_message(message=self.message)

        # 7. Get banner message by configurations api
        configurations = self.configurations.get_configurations()

        # 8. Check banner message by configurations api
        config_banner_message = configurations.banner_message
        self.assertEqual(config_banner_message.editable, True)
        self.checkBannerMseeage(config_banner_message.value)

        # 9. Get banner message by system info api
        system_info = self.system_info.get_system_info()

        # 10. Check banner message by system info api
        self.checkBannerMseeage(system_info.banner_message)


    def checkBannerMseeage(self, banner_mseeage):
        if self.message == "":
            self.assertEqual(banner_mseeage, "")
        else:
            self.assertEqual(banner_mseeage["message"], self.message)
            self.assertEqual(banner_mseeage["type"], self.message_type)
            self.assertEqual(banner_mseeage["closable"], self.closable)
            self.assertEqual(banner_mseeage["fromDate"], self.from_date)
            self.assertEqual(banner_mseeage["toDate"], self.to_date)


if __name__ == '__main__':
    unittest.main()
