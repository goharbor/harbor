from __future__ import absolute_import

import unittest
import time

from testutils import ADMIN_CLIENT
from testutils import TEARDOWN
from testutils import harbor_server
from library.repository import push_special_image_to_project

from library.retention import Retention
from library.project import Project
from library.repository import Repository
from library.user import User
from library.system import System


class TestProjects(unittest.TestCase):
    """
    Test case:
        Tag Retention
    Setup:
        Create Project test-retention
        Push image test1:1.0, test1:2.0, test1:3.0,latest, test2:1.0, test2:latest, test3:1.0, test4:1.0
    Test Steps:
        1. Create Retention Policy
        2. Add rule "For the repositories matching **, retain always with tags matching latest*"
        3. Add rule "For the repositories matching test3*, retain always with tags matching **"
        4. Dry run, check execution and tasks
        5. Real run, check images retained
    Tear Down:
        1. Delete project test-retention
    """
    @classmethod
    def setUpClass(self):
        self.user = User()
        self.system = System()
        self.repo= Repository()
        self.project = Project()
        self.retention=Retention()

    def testTagRetention(self):
        user_ra_password = "Aa123456"
        user_ra_id, user_ra_name = self.user.create_user(user_password=user_ra_password, **ADMIN_CLIENT)
        print("Created user: %s, id: %s" % (user_ra_name, user_ra_id))
        TestProjects.USER_RA_CLIENT = dict(endpoint=ADMIN_CLIENT["endpoint"],
                                   username=user_ra_name,
                                   password=user_ra_password)
        TestProjects.user_ra_id = int(user_ra_id)

        TestProjects.project_src_repo_id, project_src_repo_name = self.project.create_project(metadata = {"public": "false"}, **TestProjects.USER_RA_CLIENT)

        # Push image test1:1.0, test1:2.0, test1:3.0,latest, test2:1.0, test2:latest, test3:1.0
        push_special_image_to_project(project_src_repo_name, harbor_server, user_ra_name, user_ra_password, "test1", ['1.0'])
        push_special_image_to_project(project_src_repo_name, harbor_server, user_ra_name, user_ra_password, "test1", ['2.0'])
        push_special_image_to_project(project_src_repo_name, harbor_server, user_ra_name, user_ra_password, "test1", ['3.0','latest'])
        push_special_image_to_project(project_src_repo_name, harbor_server, user_ra_name, user_ra_password, "test2", ['1.0'])
        push_special_image_to_project(project_src_repo_name, harbor_server, user_ra_name, user_ra_password, "test2", ['latest'])
        push_special_image_to_project(project_src_repo_name, harbor_server, user_ra_name, user_ra_password, "test3", ['1.0'])
        push_special_image_to_project(project_src_repo_name, harbor_server, user_ra_name, user_ra_password, "test4", ['1.0'])

        resp=self.repo.get_repository(TestProjects.project_src_repo_id, **TestProjects.USER_RA_CLIENT)
        self.assertEqual(len(resp), 4)

        # Create Retention Policy
        retention_id = self.retention.create_retention_policy(TestProjects.project_src_repo_id, selector_repository="**", selector_tag="latest*", expect_status_code = 201, **TestProjects.USER_RA_CLIENT)

        # Add rule
        self.retention.update_retention_add_rule(retention_id,selector_repository="test3*", selector_tag="**", expect_status_code = 200, **TestProjects.USER_RA_CLIENT)

        # Dry run
        self.retention.trigger_retention_policy(retention_id, dry_run=True, **TestProjects.USER_RA_CLIENT)
        time.sleep(10)
        resp=self.retention.get_retention_executions(retention_id, **TestProjects.USER_RA_CLIENT)
        self.assertTrue(len(resp)>0)
        execution=resp[0]
        resp=self.retention.get_retention_exec_tasks(retention_id,execution.id, **TestProjects.USER_RA_CLIENT)
        self.assertEqual(len(resp), 4)
        resp=self.retention.get_retention_exec_task_log(retention_id,execution.id,resp[0].id, **TestProjects.USER_RA_CLIENT)
        print(resp)

        # Real run
        self.retention.trigger_retention_policy(retention_id, dry_run=False, **TestProjects.USER_RA_CLIENT)
        time.sleep(10)
        resp=self.retention.get_retention_executions(retention_id, **TestProjects.USER_RA_CLIENT)
        self.assertTrue(len(resp)>1)
        execution=resp[0]
        resp=self.retention.get_retention_exec_tasks(retention_id,execution.id, **TestProjects.USER_RA_CLIENT)
        self.assertEqual(len(resp), 4)
        resp=self.retention.get_retention_exec_task_log(retention_id,execution.id,resp[0].id, **TestProjects.USER_RA_CLIENT)
        print(resp)
        # TODO As the repository isn't deleted when no tags left anymore
        # TODO we should check the artifact/tag count here
        # resp=self.repo.get_repository(TestProjects.project_src_repo_id, **TestProjects.USER_RA_CLIENT)
        # self.assertEqual(len(resp), 3)


    @classmethod
    def tearDownClass(self):
        print "Case completed"

    # TODO delete_repoitory will fail when no tags left anymore
    # @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    # def test_ClearData(self):
    #     resp=self.repo.get_repository(TestProjects.project_src_repo_id, **TestProjects.USER_RA_CLIENT)
    #     for repo in resp:
    #         self.repo.delete_repoitory(repo.name, **TestProjects.USER_RA_CLIENT)
    #     self.project.delete_project(TestProjects.project_src_repo_id, **TestProjects.USER_RA_CLIENT)
    #     self.user.delete_user(TestProjects.user_ra_id, **ADMIN_CLIENT)


if __name__ == '__main__':
    unittest.main()
