from __future__ import absolute_import
import time
import unittest

from testutils import ADMIN_CLIENT, harbor_server, suppress_urllib3_warning
from library.user import User
from library.configurations import Configurations
from library.project import Project
from library.artifact import Artifact
from library.scan import Scan
from library.repository import push_self_build_image_to_project

class TestRetainImageLastPullTime(unittest.TestCase, object):

    @suppress_urllib3_warning
    def setUp(self):
        self.user = User()
        self.conf = Configurations()
        self.project = Project()
        self.artifact = Artifact()
        self.scan = Scan()
        self.image = "alpine"
        self.tag = "latest"
        self.default_time = "1-01-01 00:00:00"

    def testRetainImageLastPullTime(self):
        """
        Test case:
            RetainImageLastPullTime
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new private project(PA) by user(UA);
            3. Push a new image(IA) in project(PA) by user(UA);
            4. Enable the retain image last pull time on scanning;
            5. Scan image(IA);
            6. Check the last pull time of image(IA) is default time;
            7. Disable the retain image last pull time on scanning;
            8. Scan image(IA);
            9. Check the last pull time of image(IA) is not default time;
        """
        url = ADMIN_CLIENT["endpoint"]
        user_password = "Aa123456"
        # 1. Create a new user(UA);
        _, user_name = self.user.create_user(user_password=user_password, **ADMIN_CLIENT)
        USER_CLIENT = dict(endpoint = url, username = user_name, password = user_password, with_accessory = True)
        # 2. Create a new private project(PA) by user(UA);
        _, project_name = self.project.create_project(metadata = {"public": "false"}, **USER_CLIENT)
        # 3. Push a new image(IA) in repository(RA) by user(UA);
        push_self_build_image_to_project(project_name, harbor_server, user_name, user_password, self.image, self.tag)
        # 4. Enable the retain image last pull time on scanning;
        self.conf.set_configurations_of_retain_image_last_pull_time(True)
        # 5. Scan image(IA);
        self.scan.scan_artifact(project_name, self.image, self.tag, **USER_CLIENT)
        time.sleep(15)
        # 6. Check the last pull time of image(IA) is default time;
        artifact_info = self.artifact.get_reference_info(project_name, self.image, self.tag, **USER_CLIENT)
        self.assertEqual(artifact_info.pull_time.strftime("%Y-%m-%d %H:%M:%S"), self.default_time)
        # 7. Disable the retain image last pull time on scanning;
        self.conf.set_configurations_of_retain_image_last_pull_time(False)
        # 8. Scan image(IA);
        self.scan.scan_artifact(project_name, self.image, self.tag, **USER_CLIENT)
        # 9. Check the last pull time of image(IA) is not default time;
        pull_time = self.default_time
        for _ in range(6):
            artifact_info = self.artifact.get_reference_info(project_name, self.image, self.tag, **USER_CLIENT)
            pull_time = artifact_info.pull_time.strftime("%Y-%m-%d %H:%M:%S")
            if pull_time != self.default_time:
                break
            else:
                time.sleep(5)
        self.assertNotEqual(pull_time, self.default_time)


if __name__ == '__main__':
    unittest.main()
