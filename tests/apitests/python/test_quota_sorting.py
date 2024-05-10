from __future__ import absolute_import

import unittest

from testutils import harbor_server, suppress_urllib3_warning
from testutils import TEARDOWN
from testutils import ADMIN_CLIENT
from library.project import Project
from library.user import User
from library.repository import Repository
from library.repository import push_self_build_image_to_project
from library.quota_sorting import QuotaSorting

class TestQuotaSorting(unittest.TestCase):
    @suppress_urllib3_warning
    def setUp(self):
        self.project = Project()
        self.user = User()
        self.repo = Repository()
        self.quota_sorting = QuotaSorting()

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def tearDown(self):
        #1. Delete repository(RA) by user(UA);
        self.repo.delete_repository(TestQuotaSorting.project_1_name, TestQuotaSorting.repo_name_1.split('/')[1], **TestQuotaSorting.user_global_client)
        self.repo.delete_repository(TestQuotaSorting.project_2_name, TestQuotaSorting.repo_name_2.split('/')[1], **TestQuotaSorting.user_global_client)

        #2. Delete project(PA);
        self.project.delete_project(TestQuotaSorting.project_1_id, **TestQuotaSorting.user_global_client)
        self.project.delete_project(TestQuotaSorting.project_2_id, **TestQuotaSorting.user_global_client)

        #3. Delete user(UA);
        self.user.delete_user(TestQuotaSorting.user_001_id, **ADMIN_CLIENT)

    def testQuotaSorting(self):
        """
        Test case:
            Quota Sorting
        Test step and expected result:
            1. Create a new user(UA);
            2. Create two new private projects(PA) by user(UA);
            3. Push two images to projects(PA) respectively by user(UA)
            4. Get quota list with sort=used.storage, the used storage should be in ascending order
            5. Get quota list with sort=-used.storage, the used storage should be in descending order
        Tear down:
            1. Delete repository(RA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA);
        """
        url = ADMIN_CLIENT["endpoint"]
        user_001_password = "Aa123456"
        global_admin_client = dict(endpoint=ADMIN_CLIENT["endpoint"], username=ADMIN_CLIENT["username"], passwor=ADMIN_CLIENT["password"], reference="project")

        #1. Create user-001
        TestQuotaSorting.user_001_id, user_001_name = self.user.create_user(user_password=user_001_password, **ADMIN_CLIENT)
        TestQuotaSorting.user_global_client = dict(endpoint=url, username=user_001_name, password=user_001_password)

        #2. Create private project_1 and private project_2
        TestQuotaSorting.project_1_id, TestQuotaSorting.project_1_name = self.project.create_project(metadata={"public": "false"}, **TestQuotaSorting.user_global_client)
        TestQuotaSorting.project_2_id, TestQuotaSorting.project_2_name = self.project.create_project(metadata={"public": "false"}, **TestQuotaSorting.user_global_client)

        #3. Push images to project_1 and project_2 respectively
        image1 = "alpine"
        tag1 = "2.6"
        TestQuotaSorting.repo_name_1, _ = push_self_build_image_to_project(TestQuotaSorting.project_1_name, harbor_server, user_001_name, user_001_password, image1, tag1)
        image2 = "photon"
        tag2 = "2.0"
        TestQuotaSorting.repo_name_2, _ = push_self_build_image_to_project(TestQuotaSorting.project_2_name, harbor_server, user_001_name, user_001_password, image2, tag2)

        #4. Check whether quota list is in ascending order
        global_admin_client["sort"] = "used.storage"
        res_ascending = self.quota_sorting.list_quotas_with_sorting(expect_status_code=200, **global_admin_client)
        self.assertTrue(len(res_ascending) >= 2)
        for idx in range(1, len(res_ascending)):
            self.assertTrue(res_ascending[idx - 1].to_dict()["used"]["storage"] <= res_ascending[idx].to_dict()["used"]["storage"])

        #5. Check whether quota list is in descending order
        global_admin_client["sort"] = "-used.storage"
        res_descending = self.quota_sorting.list_quotas_with_sorting(expect_status_code=200, **global_admin_client)
        self.assertTrue(len(res_descending) >= 2)
        for idx in range(1, len(res_descending)):
            self.assertTrue(res_descending[idx - 1].to_dict()["used"]["storage"] >= res_descending[idx].to_dict()["used"]["storage"])


if __name__ == '__main__':
    unittest.main()