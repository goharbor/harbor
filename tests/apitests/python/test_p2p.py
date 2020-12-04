from __future__ import absolute_import


import unittest
import urllib
import sys

from testutils import ADMIN_CLIENT, TEARDOWN, harbor_server, suppress_urllib3_warning
from library.base import _random_name
from library.base import _assert_status_code
from library.project import Project
from library.user import User
from library.repository import Repository
from library.registry import Registry
from library.repository import pull_harbor_image
from library.artifact import Artifact
from library.preheat import Preheat
import library.containerd
import v2_swagger_client

class TestP2P(unittest.TestCase):
    @suppress_urllib3_warning
    def setUp(self):
        self.url = ADMIN_CLIENT["endpoint"]
        self.user_password = "Aa123456"
        self.project= Project()
        self.user= User()
        self.repo= Repository()
        self.registry = Registry()
        self.artifact = Artifact()
        self.preheat = Preheat()

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def tearDown(self):
        print("Case completed")

    def do_validate(self, registry_type):
        """
        Test case:
            Proxy Cache Image From Harbor
        Test step and expected result:
            1. Create a new registry;
            2. Create a new project;
            3. Add a new user as a member of project;
            4. Pull image from this project by docker CLI;
            5. Pull image from this project by ctr CLI;
            6. Pull manifest index from this project by docker CLI;
            7. Pull manifest from this project by ctr CLI;
            8. Image pulled by docker CLI should be cached;
            9. Image pulled by ctr CLI should be cached;
            10. Manifest index pulled by docker CLI should be cached;
            11. Manifest index pulled by ctr CLI should be cached;
        Tear down:
            1. Delete project(PA);
            2. Delete user(UA).
        """
        user_id, user_name = self.user.create_user(user_password = self.user_password, **ADMIN_CLIENT)
        USER_CLIENT=dict(with_signature = True, endpoint = self.url, username = user_name, password = self.user_password)

        #2. Create a new distribution instance;
        instance_id, instance_name = self.preheat.create_instance( **ADMIN_CLIENT)

        #This need to be removed once issue #13378 fixed.
        instance = self.preheat.get_instance(instance_name)
        print("instance:", instance)

        #2. Create a new project;
        project_id, project_name = self.project.create_project(metadata = {"public": "false"}, **USER_CLIENT)
        print("project_id:",project_id)
        print("project_name:",project_name)

        #This need to be removed once issue #13378 fixed.
        policy_id, policy_name = self.preheat.create_policy(project_name, project_id, instance.id, **USER_CLIENT)
        #policy_id, _ = self.preheat.create_policy(project_name, project_id, instance_id, **USER_CLIENT)
        policy = self.preheat.get_policy(project_name, policy_name)
        print("policy:", policy)

        policy_new = v2_swagger_client.PreheatPolicy(id = policy.id, name="policy_new_name", project_id=project_id, provider_id=instance.id,
                                    description="edit this policy",filters=r'[{"type":"repository","value":"zgila/alpine*"},{"type":"tag","value":"v1.0*"},{"type":"label","value":"release"}]',
                                    trigger=r'{"type":"scheduled","trigger_setting":{"cron":"0 8 * * * *"}}', enabled=False)

        self.preheat.update_policy(project_name, policy.name, policy_new, **USER_CLIENT)

        self.preheat.delete_instance(instance.name, expect_status_code=403, **USER_CLIENT)

        self.project.delete_project(project_id, **USER_CLIENT)

        self.preheat.delete_instance(instance.name, **ADMIN_CLIENT)

    def test_create_instance(self):
        self.do_validate("harbor")

    def suite():
        suite = unittest.TestSuite(unittest.makeSuite(TestP2P))
        return suite

if __name__ == '__main__':
    result = unittest.TextTestRunner(sys.stdout, verbosity=2, failfast=True).run(TestP2P.suite())
    print("Test result:",result)
    if not result.wasSuccessful():
        raise Exception(r"P2P test failed!")

