from __future__ import absolute_import


import unittest

import library.repository
import library.helm
from testutils import ADMIN_CLIENT
from testutils import harbor_server

from testutils import TEARDOWN
from library.project import Project
from library.user import User
from library.repository import Repository
from library.artifact import Artifact

class TestProjects(unittest.TestCase):
    @classmethod
    def setUpClass(self):
        self.project= Project()
        self.user= User()
        self.artifact = Artifact(api_type='artifact')
        self.repo= Repository(api_type='repository')
        self.url = ADMIN_CLIENT["endpoint"]
        self.user_push_chart_password = "Aa123456"
        self.chart_file = "https://storage.googleapis.com/harbor-builds/helm-chart-test-files/harbor-0.2.0.tgz"
        self.archive = "harbor/"
        self.verion = "0.2.0"
        self.repo_name = "harbor_api_test"

    @classmethod
    def tearDownClass(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #1. Delete repository chart(CA) by user(UA);
        self.repo.delete_repoitory(TestProjects.project_push_chart_name, self.repo_name, **TestProjects.USER_CLIENT)

        #2. Delete project(PA);
        self.project.delete_project(TestProjects.project_push_chart_id, **TestProjects.USER_CLIENT)

        #3. Delete user(UA).
        self.user.delete_user(TestProjects.user_id, **ADMIN_CLIENT)

    def testPushChartByHelmChartCLI(self):
        """
        Test case:
            Push Chart File By Helm Chart CLI
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Push an chart(CA) to Harbor by helm3 registry/chart CLI successfully;
            4. Get chart(CA) from Harbor successfully;
            5. TO_DO: Verify this chart artifact information, like digest.
        Tear down:
            1. Delete repository chart(CA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA).
        """
        #1. Create a new user(UA);
        TestProjects.user_id, user_name = self.user.create_user(user_password = self.user_push_chart_password, **ADMIN_CLIENT)
        TestProjects.USER_CLIENT=dict(endpoint = self.url, username = user_name, password = self.user_push_chart_password)

        #2. Create a new project(PA) by user(UA);
        TestProjects.project_push_chart_id, TestProjects.project_push_chart_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_CLIENT)

        #3. Push an chart(CA) to Harbor by helm3 registry/chart CLI successfully;
        chart_cli_ret = library.helm.helm_chart_push_to_harbor(self.chart_file, self.archive,  harbor_server, TestProjects.project_push_chart_name, self.repo_name, self.verion, user_name, self.user_push_chart_password)
        print "chart_cli_ret:", chart_cli_ret

        #4. Get chart(CA) from Harbor successfully;
        artifact = self.artifact.get_reference_info(TestProjects.project_push_chart_name, self.repo_name, self.verion, **TestProjects.USER_CLIENT)
        print "artifact:", artifact

        #5. TO_DO: Verify this chart artifact information, like digest;
        self.assertEqual(artifact[0].type, 'CHART')


if __name__ == '__main__':
    unittest.main()

