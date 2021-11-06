from __future__ import absolute_import

import unittest

from testutils import ADMIN_CLIENT, suppress_urllib3_warning, harbor_server, files_directory
from testutils import TEARDOWN
from library import base
from library import helm
from library.project import Project
from library.user import User
from library.repository import Repository
from library.artifact import Artifact


class TestProjects(unittest.TestCase):

    user_id = None
    project_push_chart_id = None
    USER_CLIENT = None
    project_push_chart_name = None

    @suppress_urllib3_warning
    def setUp(self):
        self.project = Project()
        self.user = User()
        self.artifact = Artifact()
        self.repo = Repository()
        self.url = ADMIN_CLIENT["endpoint"]
        self.user_push_chart_password = "Aa123456"
        self.chart_file_name = "harbor-helm-1.7.3"
        self.chart_file_package_name = "harbor-1.7.3.tgz"
        self.chart_file_path = files_directory + "harbor-helm-1.7.3.tar.gz"
        self.version = "1.7.3"
        self.repo_name = "harbor"

    @unittest.skipIf(TEARDOWN is False, "Test data won't be erased.")
    def tearDown(self):
        # 1. Delete repository chart(CA) by user(UA);
        self.repo.delete_repository(TestProjects.project_push_chart_name, self.repo_name, **TestProjects.USER_CLIENT)

        # 2. Delete project(PA);
        self.project.delete_project(TestProjects.project_push_chart_id, **TestProjects.USER_CLIENT)

        # 3. Delete user(UA).
        self.user.delete_user(TestProjects.user_id, **ADMIN_CLIENT)

    def testPushChartByHelmChartCLI(self):
        """
        Test case:
            Push Chart File By Helm3.7 CLI
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Push an chart(CA) to Harbor by helm3.7 CLI successfully;
            4. List artifacts successfully;
            5. Get chart(CA) by reference successfully;
            6. Get addition successfully;
            7. Delete chart by reference successfully.
        Tear down:
            1. Delete repository chart(CA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA).
        """
        # 1. Create a new user(UA);
        TestProjects.user_id, user_name = self.user.create_user(user_password=self.user_push_chart_password,
                                                                **ADMIN_CLIENT)
        TestProjects.USER_CLIENT = dict(endpoint=self.url, username=user_name, password=self.user_push_chart_password)

        # 2. Create a new project(PA) by user(UA);
        TestProjects.project_push_chart_id, TestProjects.project_push_chart_name = self.project.create_project(
            metadata={"public": "false"}, **TestProjects.USER_CLIENT)

        # 3 Push an chart(CA) to Harbor by helm3.7 CLI successfully;
        command = ["tar", "zxf", self.chart_file_path]
        base.run_command(command)
        # 3.1 helm3_7_registry_login;
        helm.helm3_7_registry_login(ip=harbor_server, user=user_name, password=self.user_push_chart_password)
        # 3.2 helm3_7_package;
        helm.helm3_7_package(file_path=self.chart_file_name)
        # 3.2 helm3_7_push;
        helm.helm3_7_push(file_path=self.chart_file_package_name, ip=harbor_server,
                          project_name=TestProjects.project_push_chart_name)

        # 4. List artifacts successfully;
        artifacts = self.artifact.list_artifacts(TestProjects.project_push_chart_name, self.repo_name,
                                                 **TestProjects.USER_CLIENT)
        self.assertEqual(artifacts[0].type, 'CHART')
        self.assertEqual(artifacts[0].tags[0].name, self.version)

        # 5.1 Get chart(CA) by reference successfully;
        artifact = self.artifact.get_reference_info(TestProjects.project_push_chart_name, self.repo_name, self.version,
                                                    **TestProjects.USER_CLIENT)
        self.assertEqual(artifact.type, 'CHART')
        self.assertEqual(artifact.tags[0].name, self.version)

        # 6. Get addition successfully;
        addition_r = self.artifact.get_addition(TestProjects.project_push_chart_name, self.repo_name, self.version,
                                                "readme.md", **TestProjects.USER_CLIENT)
        self.assertIn("Helm Chart for Harbor", addition_r[0])
        addition_v = self.artifact.get_addition(TestProjects.project_push_chart_name, self.repo_name, self.version,
                                                "values.yaml", **TestProjects.USER_CLIENT)
        self.assertIn("expose", addition_v[0])

        # 7. Delete chart by reference successfully.
        self.artifact.delete_artifact(TestProjects.project_push_chart_name, self.repo_name, self.version,
                                      **TestProjects.USER_CLIENT)


if __name__ == '__main__':
    unittest.main()
