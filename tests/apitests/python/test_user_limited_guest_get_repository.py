from __future__ import absolute_import
import unittest


from testutils import ADMIN_CLIENT, suppress_urllib3_warning
from testutils import harbor_server
from testutils import admin_user
from testutils import admin_pwd
from testutils import created_project
from testutils import created_user
from testutils import TEARDOWN
from library.repository import push_self_build_image_to_project
from library.repository import Repository


class TestLimitedGuestGetRepository(unittest.TestCase):


    @suppress_urllib3_warning
    def setUp(self):
        self.repository = Repository()

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def tearDown(self):
        print("Case completed")

    def testLimitedGuestGetRepository(self):
        """
        Test case:
            Limited Guest GetRepository
        Test step and expected result:
            1. Create a new user(UA)
            2. Create a private project(PA)
            3. Add (UA) as "Limited Guest" to this (PA)
            4. Push an image to project(PA)
            5. Call the "GetRepository" API, it should return 200 status code and project_id should be as expected, and the name should be "ProjectName/ImageName"
            6. Delete repository(RA)
        """
        url = ADMIN_CLIENT["endpoint"]
        user_001_password = "Aa123456"
        # 1. Create a new user(UA)
        with created_user(user_001_password) as (user_id, user_name):
            #2. Create a new private project(PA) by user(UA);
            #3. Add user(UA) as a member of project(PA) with "Limited Guest" role;
            with created_project(metadata={"public": "false"}, user_id=user_id, member_role_id=5) as (project_id, project_name):
                #4. Push an image to project(PA) by user(UA), then check the project quota usage;
                image, tag = "goharbor/alpine", "3.10"
                push_self_build_image_to_project(project_name, harbor_server, admin_user, admin_pwd, image, tag)

                #5. Call the "GetRepository" API, it should return 200 status code and the "name" attribute is "ProjectName/ImageName"
                USER_CLIENT=dict(endpoint=url, username=user_name, password=user_001_password)
                repository_data = self.repository.get_repository(project_name, "goharbor%2Falpine", **USER_CLIENT)
                self.assertEqual(repository_data.project_id, project_id)
                self.assertEqual(repository_data.name, project_name + "/" + image)

                #6. Delete repository(RA)
                self.repository.delete_repository(project_name, "goharbor%2Falpine", **ADMIN_CLIENT)


if __name__ == '__main__':
    unittest.main()

