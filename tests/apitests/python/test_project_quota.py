from __future__ import absolute_import
import unittest

from testutils import harbor_server, created_project, created_user
from testutils import ADMIN_CLIENT
from library.repository import Repository
from library.repository import push_image_to_project
from library.system import System

class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(cls):
        cls.repo = Repository(api_type='repository')
        cls.system = System()

    @classmethod
    def tearDown(cls):
        print "Case completed"

    def testProjectQuota(self):
        """
        Test case:
            Project Quota
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new private project(PA) by user(UA);
            3. Add user(UA) as a member of project(PA) with project-admin role;
            4. Push an image to project(PA) by user(UA), then check the project quota usage;
            5. Check quota change
            6. Push the image with another tag to project(PA) by user(UA)
            7. Check quota not changed
            8. Delete repository(RA) by user(UA);
            9. Delete image, the quota should be changed to 0.
        Tear down:
            1. Delete repository(RA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA);
        """
        user_001_password = "Aa123456"

        #1. Create a new user(UA);
        with created_user(user_001_password) as (user_id, user_name):
            #2. Create a new private project(PA) by user(UA);
            #3. Add user(UA) as a member of project(PA) with project-admin role;
            with created_project(metadata={"public": "false"}, user_id=user_id) as (project_id, project_name):
                #4. Push an image to project(PA) by user(UA), then check the project quota usage; -- {"count": 1, "storage": 2791709}
                image, tag = "goharbor/alpine", "3.10"
                push_image_to_project(project_name, harbor_server, user_name, user_001_password, image, tag)

                #5. Get project quota
                quota = self.system.get_project_quota("project", project_id, **ADMIN_CLIENT)
                self.assertEqual(quota[0].used["count"], 1)
                self.assertEqual(quota[0].used["storage"], 2789002)

                #6. Push the image with another tag to project(PA) by user(UA), the check the project quota usage; -- {"count": 1, "storage": 2791709}
                push_image_to_project(project_name, harbor_server, user_name, user_001_password, image, tag)

                #7. Get project quota
                quota = self.system.get_project_quota("project", project_id, **ADMIN_CLIENT)
                self.assertEqual(quota[0].used["count"], 1)
                self.assertEqual(quota[0].used["storage"], 2789002)

                #8. Delete repository(RA) by user(UA);
                self.repo.delete_repoitory(project_name, image, **ADMIN_CLIENT)

                #9. Quota should be 0
                quota = self.system.get_project_quota("project", project_id, **ADMIN_CLIENT)
                self.assertEqual(quota[0].used["count"], 0)
                self.assertEqual(quota[0].used["storage"], 0)

if __name__ == '__main__':
    unittest.main()