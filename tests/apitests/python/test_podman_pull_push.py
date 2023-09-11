# -*- coding: utf-8 -*-

from __future__ import absolute_import
import unittest

from testutils import harbor_server, ADMIN_CLIENT, suppress_urllib3_warning
from library import podman
from library.project import Project
from library.user import User
from library.artifact import Artifact
from library.repository import push_self_build_image_to_project

class TestPodmanPullPush(unittest.TestCase):

    @suppress_urllib3_warning
    def setUp(self):
        self.project= Project()
        self.user= User()
        self.artifact = Artifact()
        self.image = "image_test"
        self.tag = "v1"
        self.source_image = "ghcr.io/goharbor/harbor-core"
        self.source_tag = "v2.8.2"

    def testPodman(self):
        """
        Test case:
            Podman pull and push
        Test step and expected result:
            1. Create a new user;
            2. Create a new project by user;
            3. Push a new image in project by user;
            4. Podman login harbor;
            5. Podman pull image from project(PA) by user;
            6. Podman pull soure image;
            7. Podman push soure image to project by user;
            8. Verify the image;
            9. Podman logout harbor;
        """
        url = ADMIN_CLIENT["endpoint"]
        user_password = "Aa123456"

        # 1. Create user(UA)
        _, user_name = self.user.create_user(user_password = user_password, **ADMIN_CLIENT)
        user_client = dict(endpoint = url, username = user_name, password = user_password, with_accessory = True)

        # 2. Create private project(PA) by user(UA)
        _, project_name = self.project.create_project(metadata = {"public": "false"}, **user_client)

        # 3. Push a new image(IA) in project(PA) by user(UA)
        push_self_build_image_to_project(project_name, harbor_server, user_name, user_password, self.image, self.tag)

        # 4. Podman login harbor
        podman.login(harbor_server, user_name, user_password)

        # 5. Podman pull image from project(PA) by user
        podman.pull("{}/{}/{}:{}".format(harbor_server, project_name, self.image, self.tag))

        # 6. Podman pull soure image
        podman.pull("{}:{}".format(self.source_image, self.source_tag))

        # 7. Podman push soure image to project by user
        podman.push("{}:{}".format(self.source_image, self.source_tag), "{}/{}/{}:{}".format(harbor_server, project_name, self.image, self.tag))

        # 8. Verify the image
        image_info = self.artifact.get_reference_info(project_name, self.image, self.tag, **user_client)
        self.assertIsNotNone(image_info)
        self.assertIsNotNone(image_info.digest)
        self.assertEqual(len(image_info.tags), 1)
        self.assertEqual(image_info.tags[0].name, self.tag)

        # 9. Podman logout harbor
        podman.logout(harbor_server)

if __name__ == '__main__':
    unittest.main()
