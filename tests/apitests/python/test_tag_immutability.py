from __future__ import absolute_import


import unittest
import sys

from testutils import ADMIN_CLIENT, TEARDOWN, suppress_urllib3_warning
from testutils import harbor_server
from library.project import Project
from library.user import User
from library.repository import Repository
from library.registry import Registry
from library.artifact import Artifact
from library.tag_immutability import Tag_Immutability
from library.repository import push_special_image_to_project

class TestTagImmutability(unittest.TestCase):
    @suppress_urllib3_warning
    def setUp(self):
        self.url = ADMIN_CLIENT["endpoint"]
        self.user_password = "Aa123456"
        self.project= Project()
        self.user= User()
        self.repo= Repository()
        self.registry = Registry()
        self.artifact = Artifact()
        self.tag_immutability = Tag_Immutability()
        self.project_id, self.project_name, self.user_id, self.user_name = [None] * 4
        self.user_id, self.user_name = self.user.create_user(user_password = self.user_password, **ADMIN_CLIENT)
        self.USER_CLIENT = dict(with_signature = True, with_immutable_status = True, endpoint = self.url, username = self.user_name, password = self.user_password)
        self.exsiting_rule = dict(selector_repository="rel*", selector_tag="v2.*")
        self.project_id, self.project_name = self.project.create_project(metadata = {"public": "false"}, **self.USER_CLIENT)

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def tearDown(self):
        print("Case completed")

    def check_tag_immutability(self, artifact, tag_name, status = True):
        for tag in artifact.tags:
            if tag.name == tag_name:
                self.assertTrue(tag.immutable == status)
                return
        raise Exception("No tag {} found in artifact {}".format(tag, artifact))

    def test_disability_of_rules(self):
        """
        Test case:
            Test Disability Of Rules
        Test step and expected result:
            1. Create a new project;
            2. Push image A to the project with 2 tags A and B;
            3. Create a disabled rule matched image A with tag A;
            4. Both tags of image A should not be immutable;
            5. Enable this rule;
            6. image A with tag A should be immutable.
        """
        image_a = dict(name="image_disability_a", tag1="latest", tag2="6.2.2")

        #1. Create a new project;
        project_id, project_name = self.project.create_project(metadata = {"public": "false"}, **self.USER_CLIENT)

        #2. Push image A to the project with 2 tags;
        push_special_image_to_project(project_name, harbor_server, self.user_name, self.user_password, image_a["name"], [image_a["tag1"], image_a["tag2"]])

        #3. Create a disabled rule matched image A;
        rule_id = self.tag_immutability.create_rule(project_id, disabled = True, selector_repository=image_a["name"], selector_tag=str(image_a["tag1"])[0:2] + "*", **self.USER_CLIENT)

        #4. Both tags of image A should not be immutable;
        artifact_a = self.artifact.get_reference_info(project_name, image_a["name"], image_a["tag2"], **self.USER_CLIENT)
        print("[test_disability_of_rules] - artifact:{}".format(artifact_a))
        self.assertTrue(artifact_a)
        self.check_tag_immutability(artifact_a, image_a["tag1"], status = False)
        self.check_tag_immutability(artifact_a, image_a["tag2"], status = False)

        #5. Enable this rule;
        self.tag_immutability.update_tag_immutability_policy_rule(project_id, rule_id, disabled = False, **self.USER_CLIENT)

        #6. image A with tag A should be immutable.
        artifact_a = self.artifact.get_reference_info(project_name, image_a["name"], image_a["tag2"], **self.USER_CLIENT)
        print("[test_disability_of_rules] - artifact:{}".format(artifact_a))
        self.assertTrue(artifact_a)
        self.check_tag_immutability(artifact_a, image_a["tag1"], status = True)
        self.check_tag_immutability(artifact_a, image_a["tag2"], status = False)

    def test_artifact_and_repo_is_undeletable(self):
        """
        Test case:
            Test Artifact And Repo is Undeleteable
        Test step and expected result:
            1. Create a new project;
            2. Push image A to the project with 2 tags A and B;
            3. Create a enabled rule matched image A with tag A;
            4. Tag A should be immutable;
            5. Artifact is undeletable;
            6. Repository is undeletable.
        """
        image_a = dict(name="image_repo_undeletable_a", tag1="latest", tag2="1.3.2")

        #1. Create a new project;
        project_id, project_name = self.project.create_project(metadata = {"public": "false"}, **self.USER_CLIENT)

        #2. Push image A to the project with 2 tags A and B;
        push_special_image_to_project(project_name, harbor_server, self.user_name, self.user_password, image_a["name"], [image_a["tag1"], image_a["tag2"]])

        #3. Create a enabled rule matched image A with tag A;
        self.tag_immutability.create_rule(project_id, selector_repository=image_a["name"], selector_tag=str(image_a["tag1"])[0:2] + "*", **self.USER_CLIENT)

        #4. Tag A should be immutable;
        artifact_a = self.artifact.get_reference_info(project_name, image_a["name"], image_a["tag2"], **self.USER_CLIENT)
        print("[test_artifact_and_repo_is_undeletable] - artifact:{}".format(artifact_a))
        self.assertTrue(artifact_a)
        self.check_tag_immutability(artifact_a, image_a["tag1"], status = True)
        self.check_tag_immutability(artifact_a, image_a["tag2"], status = False)

        #5. Artifact is undeletable;
        self.artifact.delete_artifact(project_name, image_a["name"], image_a["tag1"], expect_status_code = 412,expect_response_body = "configured as immutable, cannot be deleted", **self.USER_CLIENT)

        #6. Repository is undeletable.
        self.repo.delete_repository(project_name, image_a["name"], expect_status_code = 412, expect_response_body = "configured as immutable, cannot be deleted", **self.USER_CLIENT)

    def test_tag_is_undeletable(self):
        """
        Test case:
            Test Tag is Undeleteable
        Test step and expected result:
            1. Push image A to the project with 2 tags A and B;
            2. Create a enabled rule matched image A with tag A;
            3. Tag A should be immutable;
            4. Tag A is undeletable;
            5. Tag B is deletable.
        """
        image_a = dict(name="image_undeletable_a", tag1="latest", tag2="9.3.2")

        #1. Push image A to the project with 2 tags A and B;
        push_special_image_to_project(self.project_name, harbor_server, self.user_name, self.user_password, image_a["name"], [image_a["tag1"], image_a["tag2"]])

        #2. Create a enabled rule matched image A with tag A;
        self.tag_immutability.create_rule(self.project_id, selector_repository=image_a["name"], selector_tag=str(image_a["tag2"])[0:2] + "*", **self.USER_CLIENT)

        #3. Tag A should be immutable;
        artifact_a = self.artifact.get_reference_info(self.project_name, image_a["name"], image_a["tag2"], **self.USER_CLIENT)
        print("[test_tag_is_undeletable] - artifact:{}".format(artifact_a))
        self.assertTrue(artifact_a)
        self.check_tag_immutability(artifact_a, image_a["tag2"], status = True)

        #4. Tag A is undeletable;
        self.artifact.delete_tag(self.project_name, image_a["name"], image_a["tag1"], image_a["tag2"], expect_status_code = 412, **self.USER_CLIENT)

        #5. Tag B is deletable.
        self.artifact.delete_tag(self.project_name, image_a["name"], image_a["tag1"], image_a["tag1"], **self.USER_CLIENT)

    def test_image_is_unpushable(self):
        """
        Test case:
            Test Image is Unpushable
        Test step and expected result:
            1. Create a new project;
            2. Push image A to the project with 2 tags A and B;
            3. Create a enabled rule matched image A with tag A;
            4. Tag A should be immutable;
            5. Can not push image with the same image name and with the same tag name.
        """
        image_a = dict(name="image_unpushable_a", tag1="latest", tag2="1.3.2")

        #1. Create a new project;
        project_id, project_name = self.project.create_project(metadata = {"public": "false"}, **self.USER_CLIENT)

        #2. Push image A to the project with 2 tags A and B;
        push_special_image_to_project(project_name, harbor_server, self.user_name, self.user_password, image_a["name"], [image_a["tag1"], image_a["tag2"]])

        #3. Create a enabled rule matched image A with tag A;
        self.tag_immutability.create_rule(project_id, selector_repository=image_a["name"], selector_tag=str(image_a["tag1"])[0:2] + "*", **self.USER_CLIENT)

        #4. Tag A should be immutable;
        artifact_a = self.artifact.get_reference_info(project_name, image_a["name"], image_a["tag2"], **self.USER_CLIENT)
        print("[test_image_is_unpushable] - artifact:{}".format(artifact_a))
        self.assertTrue(artifact_a)
        self.check_tag_immutability(artifact_a, image_a["tag1"], status = True)
        self.check_tag_immutability(artifact_a, image_a["tag2"], status = False)

        #5. Can not push image with the same image name and with the same tag name.
        push_special_image_to_project(project_name, harbor_server, self.user_name, self.user_password, image_a["name"], [image_a["tag1"]], size=10
                                      , expected_error_message = "configured as immutable")

    def test_copy_disability(self):
        """
        Test case:
            Test Copy Disability
        Test step and expected result:
            1. Create 2 projects;
            2. Push image A with tag A and B to project A, push image B which has the same image name and tag name to project B;
            3. Create a enabled rule matched image A with tag A;
            4. Tag A should be immutable;
            5. Can not copy artifact from project A to project B with the same repository name.
        """
        image_a = dict(name="image_copy_disability_a", tag1="latest", tag2="1.3.2")

        #1. Create 2 projects;
        project_id, project_name = self.project.create_project(metadata = {"public": "false"}, **self.USER_CLIENT)
        _, project_name_src = self.project.create_project(metadata = {"public": "false"}, **self.USER_CLIENT)

        #2. Push image A with tag A and B to project A, push image B which has the same image name and tag name to project B;
        push_special_image_to_project(project_name, harbor_server, self.user_name, self.user_password, image_a["name"], [image_a["tag1"], image_a["tag2"]])
        push_special_image_to_project(project_name_src, harbor_server, self.user_name, self.user_password, image_a["name"], [image_a["tag1"], image_a["tag2"]])

        #3. Create a enabled rule matched image A with tag A;
        self.tag_immutability.create_rule(project_id, selector_repository=image_a["name"], selector_tag=str(image_a["tag1"])[0:2] + "*", **self.USER_CLIENT)

        #4. Tag A should be immutable;
        artifact_a = self.artifact.get_reference_info(project_name, image_a["name"], image_a["tag2"], **self.USER_CLIENT)
        print("[test_copy_disability] - artifact:{}".format(artifact_a))
        self.assertTrue(artifact_a)
        self.check_tag_immutability(artifact_a, image_a["tag1"], status = True)
        self.check_tag_immutability(artifact_a, image_a["tag2"], status = False)

        #5. Can not copy artifact from project A to project B with the same repository name.
        artifact_a_src = self.artifact.get_reference_info(project_name_src, image_a["name"], image_a["tag2"], **self.USER_CLIENT)
        print("[test_copy_disability] - artifact_a_src:{}".format(artifact_a_src))
        self.artifact.copy_artifact(project_name, image_a["name"], project_name_src+"/"+ image_a["name"] + "@" + artifact_a_src.digest, expect_status_code=412, expect_response_body = "configured as immutable, cannot be updated", **self.USER_CLIENT)

    #def test_replication_disability(self):
    #    pass

    def test_priority_of_rules(self):
        """
        Test case:
            Test Priority Of Rules(excluding rule will not affect matching rule)
        Test step and expected result:
            1. Push image A, B and C, image A has only 1 tag named tag1;
            2. Create a matching rule that matches image A and tag named tag2 which is not exist;
            3. Create a excluding rule to exlude image A and B;
            4. Add a tag named tag2 to image A, tag2 should be immutable;
            5. Tag2 should be immutable;
            6. All tags in image B should be immutable;
            7. All tags in image C should not be immutable;
            8. Disable all rules.
        """
        image_a = dict(name="image_priority_a", tag1="latest", tag2="6.3.2")
        image_b = dict(name="image_priority_b", tag1="latest", tag2="0.12.0")
        image_c = dict(name="image_priority_c", tag1="latest", tag2="3.12.0")

        #1. Push image A, B and C, image A has only 1 tag named tag1;
        push_special_image_to_project(self.project_name, harbor_server, self.user_name, self.user_password, image_a["name"], [image_a["tag1"]])
        push_special_image_to_project(self.project_name, harbor_server, self.user_name, self.user_password, image_b["name"], [image_b["tag1"],image_b["tag2"]])
        push_special_image_to_project(self.project_name, harbor_server, self.user_name, self.user_password, image_c["name"], [image_c["tag1"],image_c["tag2"]])

        #2. Create a matching rule that matches image A and tag named tag2 which is not exist;
        rule_id_1 = self.tag_immutability.create_rule(self.project_id, selector_repository=image_a["name"], selector_tag=image_a["tag2"], **self.USER_CLIENT)

        #3. Create a excluding rule to exlude image A and B;
        rule_id_2 = self.tag_immutability.create_rule(self.project_id, selector_repository_decoration = "repoExcludes",
                                          selector_repository="{image_priority_a,image_priority_b}", selector_tag="**", **self.USER_CLIENT)

        #4. Add a tag named tag2 to image A, tag2 should be immutable;
        self.artifact.create_tag(self.project_name, image_a["name"], image_a["tag1"], image_a["tag2"], **self.USER_CLIENT)

        #5. Tag2 should be immutable;
        artifact_a = self.artifact.get_reference_info(self.project_name, image_a["name"], image_a["tag2"], **self.USER_CLIENT)
        print("[test_priority_of_rules] - artifact:{}".format(artifact_a))
        self.assertTrue(artifact_a)
        self.check_tag_immutability(artifact_a, image_a["tag2"], status = True)
        self.check_tag_immutability(artifact_a, image_a["tag1"], status = False)

        #6. All tags in image B should be immutable;
        artifact_b = self.artifact.get_reference_info(self.project_name, image_b["name"], image_b["tag2"], **self.USER_CLIENT)
        print("[test_priority_of_rules] - artifact:{}".format(artifact_b))
        self.assertTrue(artifact_b)
        self.check_tag_immutability(artifact_b, image_b["tag2"], status = False)
        self.check_tag_immutability(artifact_b, image_b["tag1"], status = False)

        #7. All tags in image C should not be immutable;
        artifact_c = self.artifact.get_reference_info(self.project_name, image_c["name"], image_c["tag2"], **self.USER_CLIENT)
        print("[test_priority_of_rules] - artifact:{}".format(artifact_c))
        self.assertTrue(artifact_c)
        self.check_tag_immutability(artifact_c, image_c["tag2"], status = True)
        self.check_tag_immutability(artifact_c, image_c["tag1"], status = True)

        #8. Disable all rules.
        self.tag_immutability.update_tag_immutability_policy_rule(self.project_id, rule_id_1, disabled = True, **self.USER_CLIENT)
        self.tag_immutability.update_tag_immutability_policy_rule(self.project_id, rule_id_2, disabled = True, **self.USER_CLIENT)

    def test_add_exsiting_rule(self):
        """
        Test case:
            Test Priority Of Rules(excluding rule will not affect matching rule)
        Test step and expected result:
            1. Push image A and B with no tag;
            2. Create a immutability policy rule A;
            3. Fail to create rule B which has the same config as rule A;
        """
        self.tag_immutability.create_tag_immutability_policy_rule(self.project_id, **self.exsiting_rule, **self.USER_CLIENT)
        self.tag_immutability.create_tag_immutability_policy_rule(self.project_id, **self.exsiting_rule, expect_status_code = 409, **self.USER_CLIENT)

if __name__ == '__main__':
    suite = unittest.TestSuite(unittest.makeSuite(TestTagImmutability))
    result = unittest.TextTestRunner(sys.stdout, verbosity=2, failfast=True).run(suite)
    if not result.wasSuccessful():
        raise Exception(r"Tag immutability test failed: {}".format(result))

