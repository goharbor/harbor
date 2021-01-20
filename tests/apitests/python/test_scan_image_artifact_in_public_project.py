from __future__ import absolute_import
import unittest

from testutils import harbor_server, TEARDOWN, suppress_urllib3_warning
from testutils import created_user, created_project
from library.artifact import Artifact
from library.repository import Repository, push_self_build_image_to_project
from library.scan import Scan


class TestScanImageInPublicProject(unittest.TestCase):
    @suppress_urllib3_warning
    def setUp(self):
        self.artifact = Artifact()
        self.repo = Repository()
        self.scan = Scan()

    @unittest.skipIf(TEARDOWN == False, "Test data won't be erased.")
    def tearDown(self):
        print("Case completed")

    def testScanImageInPublicProject(self):
        """
        Test case:
            Scan An Image Artifact In Public Project
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new public project(PA) by user(UA);
            3. Add user(UA) as a member of project(PA) with project-admin role;
            4. Create a new repository(RA) and tag(TA) in project(PA) by user(UA);
            5. Send scan image command without credential (anonymous), the API response should be 401;
            6. Create a new user(UB) which is non member of the project(PA);
            7. Send scan image command with credential of the new created user(UB), the API response should be 403;
            8. Delete user(UB);
            9. Send scan image command with credential of the user(UA) and get tag(TA) information to check scan result, it should be finished;
            10. Delete repository(RA) by user(UA);
            11. Delete project(PA);
            12. Delete user(UA);
        """
        password = 'Aa123456' # nosec
        with created_user(password) as (user_id, username):
            with created_project(metadata={"public": "true"}, user_id=user_id) as (_, project_name):
                image, src_tag = "docker", "1.13"
                full_name, tag = push_self_build_image_to_project(project_name, harbor_server, username, password, image, src_tag)

                repo_name = full_name.split('/')[1]

                # scan image with anonymous user
                self.scan.scan_artifact(project_name, repo_name, tag, expect_status_code=401, username=None, password=None)

                with created_user(password) as (_, username1):
                    # scan image with non project memeber
                    self.scan.scan_artifact(project_name, repo_name, tag, expect_status_code=403, username=username1, password=password)

                self.scan.scan_artifact(project_name, repo_name, tag, username=username, password=password)
                self.artifact.check_image_scan_result(project_name, image, tag, username=username, password=password, with_scan_overview=True)

                self.repo.delete_repository(project_name, repo_name)


if __name__ == '__main__':
    unittest.main()
