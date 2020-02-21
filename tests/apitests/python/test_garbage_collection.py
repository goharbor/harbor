from __future__ import absolute_import

import unittest

from testutils import ADMIN_CLIENT
from testutils import TEARDOWN
from library.user import User
from library.system import System
from library.project import Project
from library.repository import Repository
from library.repository import push_image_to_project
from testutils import harbor_server
from library.base import _assert_status_code

class TestProjects(unittest.TestCase):
    @classmethod
    def setUp(self):
        system = System()
        self.system= system

        project = Project()
        self.project= project

        user = User()
        self.user= user

        repo = Repository(api_type='repository')
        self.repo= repo

    @classmethod
    def tearDown(self):
        print "Case completed"

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def test_ClearData(self):
        #2. Delete project(PA);
        self.project.delete_project(TestProjects.project_gc_id, **TestProjects.USER_GC_CLIENT)

        #3. Delete user(UA);
        self.user.delete_user(TestProjects.user_gc_id, **ADMIN_CLIENT)

    def testGarbageCollection(self):
        """
        Test case:
            Garbage Collection
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Push a new image(IA) in project(PA) by admin;
            4. Delete repository(RA) by user(UA);
            5. Get repository by user(UA), it should get nothing;
            6. Tigger garbage collection operation;
            7. Check garbage collection job was finished;
            8. Get garbage collection log, check there is number of files was deleted.
        Tear down:
            1. Delete project(PA);
            2. Delete user(UA).
        """
        url = ADMIN_CLIENT["endpoint"]
        admin_name = ADMIN_CLIENT["username"]
        admin_password = ADMIN_CLIENT["password"]
        user_gc_password = "Aa123456"

        #1. Create a new user(UA);
        TestProjects.user_gc_id, user_gc_name = self.user.create_user(user_password = user_gc_password, **ADMIN_CLIENT)

        TestProjects.USER_GC_CLIENT=dict(endpoint = url, username = user_gc_name, password = user_gc_password)

        #2. Create a new project(PA) by user(UA);
        TestProjects.project_gc_id, TestProjects.project_gc_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_GC_CLIENT)

        #3. Push a new image(IA) in project(PA) by admin;
        repo_name, _ = push_image_to_project(TestProjects.project_gc_name, harbor_server, admin_name, admin_password, "tomcat", "latest")

        #4. Delete repository(RA) by user(UA);
        self.repo.delete_repoitory(TestProjects.project_gc_name, repo_name.split('/')[1], **TestProjects.USER_GC_CLIENT)

        #5. Get repository by user(UA), it should get nothing;
        repo_data = self.repo.get_repository(TestProjects.project_gc_name, **TestProjects.USER_GC_CLIENT)
        _assert_status_code(len(repo_data), 0)

        #6. Tigger garbage collection operation;
        gc_id = self.system.gc_now(**ADMIN_CLIENT)

        #7. Check garbage collection job was finished;
        self.system.validate_gc_job_status(gc_id, "finished", **ADMIN_CLIENT)

        #8. Get garbage collection log, check there is number of files was deleted.
        self.system.validate_deletion_success(gc_id, **ADMIN_CLIENT)

if __name__ == '__main__':
    unittest.main()