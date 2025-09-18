from __future__ import absolute_import

import sys
import unittest

from testutils import (
    ADMIN_CLIENT,
    TEARDOWN,
    harbor_server,
    suppress_urllib3_warning,
)
from testutils import created_user, created_project
from library.user import User
from library.project import Project
from library.robot import Robot
from library.repository import Repository
from library.artifact import Artifact
from library.repository import pull_harbor_image
from library.repository import push_self_build_image_to_project
from library.base import _assert_status_code
from library.scan import Scan
from library.label import Label
import base
import v2_swagger_client


class TestRobotAccount(unittest.TestCase):
    @suppress_urllib3_warning
    def setUp(self):
        self.project = Project()
        self.user = User()
        self.repo = Repository()
        self.artifact = Artifact()
        self.robot = Robot()
        self.scan = Scan()
        self.label = Label()

        TestRobotAccount.url = ADMIN_CLIENT["endpoint"]
        TestRobotAccount.user_ra_password = "Aa123456"
        print("setup")

    @unittest.skipIf(TEARDOWN is True, "Test data won't be erased.")
    def do_01_tearDown(self):
        # 1. Delete repository(RA) by user(UA);
        self.repo.delete_repository(
            self.project_ra_name_a,
            self.repo_name_in_project_a.split("/")[1],
            **self.USER_RA_CLIENT
        )
        self.repo.delete_repository(
            self.project_ra_name_b,
            self.repo_name_in_project_b.split("/")[1],
            **self.USER_RA_CLIENT
        )
        self.repo.delete_repository(
            self.project_ra_name_c,
            self.repo_name_in_project_c.split("/")[1],
            **self.USER_RA_CLIENT
        )
        self.repo.delete_repository(
            self.project_ra_name_a,
            self.repo_name_pa.split("/")[1],
            **self.USER_RA_CLIENT
        )

        # 2. Delete project(PA);
        self.project.delete_project(self.project_ra_id_a, **self.USER_RA_CLIENT)
        self.project.delete_project(self.project_ra_id_b, **self.USER_RA_CLIENT)
        self.project.delete_project(self.project_ra_id_c, **self.USER_RA_CLIENT)
        self.project.delete_project(self.project_ra_id_d, **self.USER_RA_CLIENT)

        # 3. Delete user(UA).
        self.user.delete_user(self.user_ra_id, **ADMIN_CLIENT)
        self.user.delete_user(self.user_ra_id_b, **ADMIN_CLIENT)

    def test_01_ProjectlevelRobotAccount(self):
        """
        Test case:
            Robot Account
        Test step and expected result:
                        1. Create user(UA);
                        2. Create private project(PA), private project(PB) and public project(PC) by user(UA);
                        3. Push image(ImagePA) to project(PA), image(ImagePB) to project(PB) and image(ImagePC) to project(PC) by user(UA);
                        4. Create a new robot account(RA) with pull and push privilege in project(PA) by user(UA);
                        5. Check robot account info, it should has both pull and push privileges;
                        6. Pull image(ImagePA) from project(PA) by robot account(RA), it must be successful;
                        7. Push image(ImageRA) to project(PA) by robot account(RA), it must be successful;
                        8. Push image(ImageRA) to project(PB) by robot account(RA), it must be not successful;
                        9. Pull image(ImagePB) from project(PB) by robot account(RA), it must be not successful;
                        10. Pull image from project(PC), it must be successful;
                        11. Push image(ImageRA) to project(PC) by robot account(RA), it must be not successful;
                        12. Update action property of robot account(RA);
                        13. Pull image(ImagePA) from project(PA) by robot account(RA), it must be not successful;
                        14. Push image(ImageRA) to project(PA) by robot account(RA), it must be not successful;
                        15. Delete robot account(RA).
            16. Create user(UB), Create public project(PD) by user(UB), user(UA) can't create robot account for project(PD).
        Tear down:
            1. Delete repository(RA) by user(UA);
            2. Delete project(PA);
            3. Delete user(UA).
        """
        image_project_a = "haproxy"
        image_project_b = "image_project_b"
        image_project_c = "httpd"
        image_robot_account = "alpine"
        tag = "latest"

        # 1. Create user(UA);"
        self.user_ra_id, user_ra_name = self.user.create_user(
            user_password=TestRobotAccount.user_ra_password, **ADMIN_CLIENT
        )
        self.USER_RA_CLIENT = dict(
            endpoint=TestRobotAccount.url,
            username=user_ra_name,
            password=TestRobotAccount.user_ra_password,
        )

        # 2. Create private project(PA), private project(PB) and public project(PC) by user(UA);
        self.project_ra_id_a, self.project_ra_name_a = self.project.create_project(
            metadata={"public": "false"}, **self.USER_RA_CLIENT
        )
        self.project_ra_id_b, self.project_ra_name_b = self.project.create_project(
            metadata={"public": "false"}, **self.USER_RA_CLIENT
        )
        self.project_ra_id_c, self.project_ra_name_c = self.project.create_project(
            metadata={"public": "true"}, **self.USER_RA_CLIENT
        )

        # 3. Push image(ImagePA) to project(PA), image(ImagePB) to project(PB) and image(ImagePC) to project(PC) by user(UA);
        self.repo_name_in_project_a, tag_a = push_self_build_image_to_project(
            self.project_ra_name_a,
            harbor_server,
            user_ra_name,
            TestRobotAccount.user_ra_password,
            image_project_a,
            tag,
        )
        self.repo_name_in_project_b, tag_b = push_self_build_image_to_project(
            self.project_ra_name_b,
            harbor_server,
            user_ra_name,
            TestRobotAccount.user_ra_password,
            image_project_b,
            tag,
        )
        self.repo_name_in_project_c, tag_c = push_self_build_image_to_project(
            self.project_ra_name_c,
            harbor_server,
            user_ra_name,
            TestRobotAccount.user_ra_password,
            image_project_c,
            tag,
        )

        # 4. Create a new robot account(RA) with pull and push privilege in project(PA) by user(UA);
        robot_id_a, robot_account_a = self.robot.create_project_robot(
            self.project_ra_name_a, 30, **self.USER_RA_CLIENT
        )
        robot_id_b, robot_account_b = self.robot.create_project_robot(
            self.project_ra_name_b, 30, **self.USER_RA_CLIENT
        )

        # 5. Check robot account info, it should has both pull and push privilege;
        data = self.robot.get_robot_account_by_id(robot_id_a, **self.USER_RA_CLIENT)
        _assert_status_code(robot_account_a.name, data.name)

        # 6. Pull image(ImagePA) from project(PA) by robot account(RA), it must be successful;
        pull_harbor_image(
            harbor_server,
            robot_account_a.name,
            robot_account_a.secret,
            self.repo_name_in_project_a,
            tag_a,
        )

        # 7. Push image(ImageRA) to project(PA) by robot account(RA), it must be successful;
        self.repo_name_pa, _ = push_self_build_image_to_project(
            self.project_ra_name_a,
            harbor_server,
            robot_account_a.name,
            robot_account_a.secret,
            image_robot_account,
            tag,
        )

        # 8. Push image(ImageRA) to project(PB) by robot account(RA), it must be not successful;
        push_self_build_image_to_project(
            self.project_ra_name_b,
            harbor_server,
            robot_account_a.name,
            robot_account_a.secret,
            image_robot_account,
            tag,
            expected_error_message="unauthorized to access repository",
        )

        # 9. Pull image(ImagePB) from project(PB) by robot account(RA), it must be not successful;
        pull_harbor_image(
            harbor_server,
            robot_account_a.name,
            robot_account_a.secret,
            self.repo_name_in_project_b,
            tag_b,
            expected_error_message="unauthorized to access repository",
        )

        # 10. Pull image from project(PC), it must be successful;
        pull_harbor_image(
            harbor_server,
            robot_account_a.name,
            robot_account_a.secret,
            self.repo_name_in_project_c,
            tag_c,
        )

        # 11. Push image(ImageRA) to project(PC) by robot account(RA), it must be not successful;
        push_self_build_image_to_project(
            self.project_ra_name_c,
            harbor_server,
            robot_account_a.name,
            robot_account_a.secret,
            image_robot_account,
            tag,
            expected_error_message="unauthorized to access repository",
        )

        # 12. Update action property of robot account(RA);"
        self.robot.disable_robot_account(robot_id_a, True, **self.USER_RA_CLIENT)

        # 13. Pull image(ImagePA) from project(PA) by robot account(RA), it must be not successful;
        pull_harbor_image(
            harbor_server,
            robot_account_a.name,
            robot_account_a.secret,
            self.repo_name_in_project_a,
            tag_a,
            expected_login_error_message="unauthorized: authentication required",
        )

        # 14. Push image(ImageRA) to project(PA) by robot account(RA), it must be not successful;
        push_self_build_image_to_project(
            self.project_ra_name_a,
            harbor_server,
            robot_account_a.name,
            robot_account_a.secret,
            image_robot_account,
            tag,
            expected_login_error_message="unauthorized: authentication required",
        )

        # 15. Delete robot account(RA).
        self.robot.delete_robot_account(robot_id_a, **self.USER_RA_CLIENT)

        # 16. Create user(UB), Create public project(PD) by user(UB), user(UA) can't create robot account for project(PD).
        self.user_ra_id_b, user_ra_name_b = self.user.create_user(
            user_password=TestRobotAccount.user_ra_password, **ADMIN_CLIENT
        )
        self.USER_RA_CLIENT_B = dict(
            endpoint=TestRobotAccount.url,
            username=user_ra_name_b,
            password=TestRobotAccount.user_ra_password,
        )
        self.project_ra_id_d, self.project_ra_name_d = self.project.create_project(
            metadata={"public": "true"}, **self.USER_RA_CLIENT_B
        )
        self.robot.create_project_robot(
            self.project_ra_name_d, 30, expect_status_code=403, **self.USER_RA_CLIENT
        )

        self.do_01_tearDown()

    def verify_repository_pushable(self, project_access_list, system_ra_client):
        for project_access in project_access_list:
            print(r"project_access:", project_access)
            if project_access["check_list"][1]:  # ---repository:push---
                push_self_build_image_to_project(
                    project_access["project_name"],
                    harbor_server,
                    system_ra_client["username"],
                    system_ra_client["password"],
                    "test_pushable" + base._random_name("repo"),
                    "v6.8.1" + base._random_name("tag"),
                )
            else:
                push_self_build_image_to_project(
                    project_access["project_name"],
                    harbor_server,
                    system_ra_client["username"],
                    system_ra_client["password"],
                    "test_unpushable" + base._random_name("repo"),
                    "v6.8.1" + base._random_name("tag"),
                    expected_error_message="unauthorized to access repository",
                )

    def verify_repository_unpushable(
        self,
        project_access_list,
        system_ra_client,
        expected_login_error_message="unauthorized: authentication required",
        expected_error_message="",
    ):
        for project_access in project_access_list:  # ---repository:push---
            push_self_build_image_to_project(
                project_access["project_name"],
                harbor_server,
                system_ra_client["username"],
                system_ra_client["password"],
                "test_unpushable" + base._random_name("repo"),
                "v6.8.1" + base._random_name("tag"),
                expected_login_error_message=expected_login_error_message,
                expected_error_message=expected_error_message,
            )

    def test_02_SystemlevelRobotAccount(self):
        """
        Test case:
            Robot Account
        Test step and expected result:
                        1. Define a number of access lists;
            2. Create the same number of private projects;
                        3. Create a system robot account has permission for those projects;
            4. Verify the system robot account has all the corresponding rights;
                        5. Disable the system robot account;
            6. Verify the system robot account has no the corresponding rights;
                        7. Enable the system robot account;
            8. Verify the system robot account has the corresponding rights;
                        9. Refresh secret for the system robot account;
            10. Verify the system robot account has no the corresponding right with the old secret already;
            11. Verify the system robot account still has the corresponding right with the new secret;
                        12. List system robot account, then add a new project to the system robot account project permission list;
            13. Verify the system robot account has the corresponding right for this new project;
            14. Edit the system robot account as removing this new project from it;
            15. Verify the system robot account has no the corresponding right for this new project;
            16. Delete this project;
            17. List system robot account successfully;
            18. Delete the system robot account;
            19. Verify the system robot account has no the corresponding right;
            20. Add a system robot account with all projects coverd;
            21. Verify the system robot account has no the corresponding right;
        """
        # 1. Define a number of access lists;

        # In this priviledge check list, make sure that each of lines and rows must
        #   contains both True and False value.
        check_list = [
            [True, True, True, False, True, False, True],
            [False, False, False, False, True, True, False],
            [True, False, True, True, False, True, True],
            [False, False, False, False, True, True, False],
        ]
        access_list_list = []
        for i in range(len(check_list)):
            access_list_list.append(self.robot.create_access_list(check_list[i]))

        # 2. Create the same number of private projects;
        robot_account_Permissions_list = []
        project_access_list = []
        for i in range(len(check_list)):
            with created_user(TestRobotAccount.user_ra_password, _teardown=False) as (
                user_id,
                username,
            ):
                with created_project(
                    metadata={"public": "false"}, user_id=user_id, _teardown=False
                ) as (project_id, project_name):
                    project_access_list.append(
                        dict(
                            project_name=project_name,
                            project_id=project_id,
                            check_list=check_list[i],
                        )
                    )
                    robot_account_Permissions = v2_swagger_client.RobotPermission(
                        kind="project",
                        namespace=project_name,
                        access=access_list_list[i],
                    )
                    robot_account_Permissions_list.append(robot_account_Permissions)

        # 3. Create a system robot account has permission for those projects;
        system_robot_account_id, system_robot_account = self.robot.create_system_robot(
            robot_account_Permissions_list, 300
        )
        print("system_robot_account:", system_robot_account)
        SYSTEM_RA_CLIENT = dict(
            endpoint=TestRobotAccount.url,
            username=system_robot_account.name,
            password=system_robot_account.secret,
        )

        # 4. Verify the system robot account has all the corresponding rights;
        for project_access in project_access_list:
            print(r"project_access:", project_access)
            if project_access["check_list"][1]:  # ---repository:push---
                push_self_build_image_to_project(
                    project_access["project_name"],
                    harbor_server,
                    SYSTEM_RA_CLIENT["username"],
                    SYSTEM_RA_CLIENT["password"],
                    "test_pushable",
                    "v6.8.1",
                )
            else:
                push_self_build_image_to_project(
                    project_access["project_name"],
                    harbor_server,
                    SYSTEM_RA_CLIENT["username"],
                    SYSTEM_RA_CLIENT["password"],
                    "test_unpushable",
                    "v6.8.1",
                    expected_error_message="unauthorized to access repository",
                )

            tag_for_del = "v1.0.0"
            repo_name, tag = push_self_build_image_to_project(
                project_access["project_name"],
                harbor_server,
                ADMIN_CLIENT["username"],
                ADMIN_CLIENT["password"],
                "test_del_artifact",
                tag_for_del,
            )
            if project_access["check_list"][0]:  # ---repository:pull---
                pull_harbor_image(
                    harbor_server,
                    SYSTEM_RA_CLIENT["username"],
                    SYSTEM_RA_CLIENT["password"],
                    repo_name,
                    tag_for_del,
                )
            else:
                pull_harbor_image(
                    harbor_server,
                    SYSTEM_RA_CLIENT["username"],
                    SYSTEM_RA_CLIENT["password"],
                    repo_name,
                    tag_for_del,
                    expected_error_message="action: pull: unauthorized to access repository",
                )

            if project_access["check_list"][2]:  # ---artifact:delete---
                self.artifact.delete_artifact(
                    project_access["project_name"],
                    repo_name.split("/")[1],
                    tag_for_del,
                    **SYSTEM_RA_CLIENT
                )
            else:
                self.artifact.delete_artifact(
                    project_access["project_name"],
                    repo_name.split("/")[1],
                    tag_for_del,
                    expect_status_code=403,
                    **SYSTEM_RA_CLIENT
                )

            repo_name, tag = push_self_build_image_to_project(
                project_access["project_name"],
                harbor_server,
                ADMIN_CLIENT["username"],
                ADMIN_CLIENT["password"],
                "test_create_tag",
                "latest_1",
            )
            self.artifact.create_tag(
                project_access["project_name"],
                repo_name.split("/")[1],
                tag,
                "for_delete",
                **ADMIN_CLIENT
            )
            if project_access["check_list"][3]:  # ---tag:create---
                self.artifact.create_tag(
                    project_access["project_name"],
                    repo_name.split("/")[1],
                    tag,
                    "1.0",
                    **SYSTEM_RA_CLIENT
                )
            else:
                self.artifact.create_tag(
                    project_access["project_name"],
                    repo_name.split("/")[1],
                    tag,
                    "1.0",
                    expect_status_code=403,
                    **SYSTEM_RA_CLIENT
                )

            if project_access["check_list"][4]:  # ---tag:delete---
                self.artifact.delete_tag(
                    project_access["project_name"],
                    repo_name.split("/")[1],
                    tag,
                    "for_delete",
                    **SYSTEM_RA_CLIENT
                )
            else:
                self.artifact.delete_tag(
                    project_access["project_name"],
                    repo_name.split("/")[1],
                    tag,
                    "for_delete",
                    expect_status_code=403,
                    **SYSTEM_RA_CLIENT
                )

            repo_name, tag = push_self_build_image_to_project(
                project_access["project_name"],
                harbor_server,
                ADMIN_CLIENT["username"],
                ADMIN_CLIENT["password"],
                "test_create_artifact_label",
                "latest_1",
            )
            # Add project level label to artifact
            label_id, _ = self.label.create_label(
                project_id=project_access["project_id"], scope="p", **ADMIN_CLIENT
            )
            if project_access["check_list"][5]:  # ---artifact-label:create---
                self.artifact.add_label_to_reference(
                    project_access["project_name"],
                    repo_name.split("/")[1],
                    tag,
                    int(label_id),
                    **SYSTEM_RA_CLIENT
                )
            else:
                self.artifact.add_label_to_reference(
                    project_access["project_name"],
                    repo_name.split("/")[1],
                    tag,
                    int(label_id),
                    expect_status_code=403,
                    **SYSTEM_RA_CLIENT
                )

            if project_access["check_list"][6]:  # ---scan:create---
                self.scan.scan_artifact(
                    project_access["project_name"],
                    repo_name.split("/")[1],
                    tag,
                    **SYSTEM_RA_CLIENT
                )
            else:
                self.scan.scan_artifact(
                    project_access["project_name"],
                    repo_name.split("/")[1],
                    tag,
                    expect_status_code=403,
                    **SYSTEM_RA_CLIENT
                )

        # 5. Disable the system robot account;
        self.robot.update_system_robot_account(
            system_robot_account_id,
            system_robot_account.name,
            robot_account_Permissions_list,
            disable=True,
            **ADMIN_CLIENT
        )

        # 6. Verify the system robot account has no the corresponding rights;
        self.verify_repository_unpushable(project_access_list, SYSTEM_RA_CLIENT)

        # 7. Enable the system robot account;
        self.robot.update_system_robot_account(
            system_robot_account_id,
            system_robot_account.name,
            robot_account_Permissions_list,
            disable=False,
            **ADMIN_CLIENT
        )

        # 8. Verify the system robot account has the corresponding rights;
        self.verify_repository_pushable(project_access_list, SYSTEM_RA_CLIENT)

        # 9. Refresh secret for the system robot account;
        new_secret = "new_secret_At_321"
        self.robot.refresh_robot_account_secret(
            system_robot_account_id, new_secret, **ADMIN_CLIENT
        )

        # 10. Verify the system robot account has no the corresponding right with the old secret already;
        self.verify_repository_unpushable(project_access_list, SYSTEM_RA_CLIENT)

        # 11. Verify the system robot account still has the corresponding right with the new secret;
        SYSTEM_RA_CLIENT["password"] = new_secret
        self.verify_repository_pushable(project_access_list, SYSTEM_RA_CLIENT)

        # 12. List system robot account, then add a new project to the system robot account project permission list;
        self.robot.list_robot(**ADMIN_CLIENT)
        project_for_del_id, project_for_del_name = self.project.create_project(
            metadata={"public": "true"}, **ADMIN_CLIENT
        )
        robot_account_Permissions = v2_swagger_client.RobotPermission(
            kind="project", namespace=project_for_del_name, access=access_list_list[0]
        )
        robot_account_Permissions_list.append(robot_account_Permissions)
        self.robot.update_system_robot_account(
            system_robot_account_id,
            system_robot_account.name,
            robot_account_Permissions_list,
            **ADMIN_CLIENT
        )
        self.robot.list_robot(**ADMIN_CLIENT)

        # 13. Verify the system robot account has the corresponding right for this new project;
        project_access_list.append(
            dict(
                project_name=project_for_del_name,
                project_id=project_for_del_id,
                check_list=[True] * 10,
            )
        )
        self.verify_repository_pushable(project_access_list, SYSTEM_RA_CLIENT)

        # 14. Edit the system robot account as removing this new project from it;
        robot_account_Permissions_list.remove(robot_account_Permissions)
        self.robot.update_system_robot_account(
            system_robot_account_id,
            system_robot_account.name,
            robot_account_Permissions_list,
            **ADMIN_CLIENT
        )
        self.robot.list_robot(**ADMIN_CLIENT)

        # 15. Verify the system robot account has no the corresponding right for this new project;
        project_access_list_for_del = [
            dict(
                project_name=project_for_del_name,
                project_id=project_for_del_id,
                check_list=[True] * 10,
            )
        ]
        self.verify_repository_unpushable(
            project_access_list_for_del,
            SYSTEM_RA_CLIENT,
            expected_login_error_message="",
            expected_error_message="action: push: unauthorized to access repository",
        )

        # 16. Delete this project;
        self.repo.clear_repositories(project_for_del_name, **ADMIN_CLIENT)
        self.project.delete_project(project_for_del_id, **ADMIN_CLIENT)

        # 17. List system robot account successfully;
        self.robot.list_robot(**ADMIN_CLIENT)

        # 18. Delete the system robot account;
        self.robot.delete_robot_account(system_robot_account_id, **ADMIN_CLIENT)

        # 19. Verify the system robot account has no the corresponding right;
        self.verify_repository_unpushable(project_access_list, SYSTEM_RA_CLIENT)

        # 20. Add a system robot account with all projects coverd;
        all_true_access_list = self.robot.create_access_list([True] * 7)
        robot_account_Permissions_list = []
        robot_account_Permissions = v2_swagger_client.RobotPermission(
            kind="project", namespace="*", access=all_true_access_list
        )
        robot_account_Permissions_list.append(robot_account_Permissions)
        _, system_robot_account_cover_all = self.robot.create_system_robot(
            robot_account_Permissions_list, 300
        )

        # 21. Verify the system robot account has no the corresponding right;
        print("system_robot_account_cover_all:", system_robot_account_cover_all)
        SYSTEM_RA_CLIENT_COVER_ALL = dict(
            endpoint=TestRobotAccount.url,
            username=system_robot_account_cover_all.name,
            password=system_robot_account_cover_all.secret,
        )
        projects = self.project.get_projects(dict(), **ADMIN_CLIENT)
        print("All projects:", projects)
        project_access_list = []
        for i in range(len(projects)):
            project_access_list.append(
                dict(
                    project_name=projects[i].name,
                    project_id=projects[i].project_id,
                    check_list=all_true_access_list,
                )
            )
        self.verify_repository_pushable(project_access_list, SYSTEM_RA_CLIENT_COVER_ALL)

    def test_03_SystemRobotCreatesProjectRobot(self):
        """
        Test case: Verify system-level robot account can create project-level robot accounts

        Test steps:
        1. Create a test project using admin credentials
        2. Create a system-level robot account with robot creation permissions for the project
        3. Use the system robot credentials to create a project-level robot account
        4. Verify the project robot was created successfully
        5. Clean up: Delete created robots and project
        """

        # Step 1: Create a test project using admin credentials
        project_id, project_name = self.project.create_project(
            metadata={"public": "false"}, **ADMIN_CLIENT
        )
        print("Created project: {} (ID: {})".format(project_name, project_id))

        # Step 2: Create system-level robot with system-level robot creation permissions
        # Define permissions: robot resource with create action at system level
        robot_access = v2_swagger_client.Access(resource="robot", action="create")
        robot_permission = v2_swagger_client.RobotPermission(
            kind="project", namespace="*", access=[robot_access]
        )

        system_robot_id, system_robot = self.robot.create_system_robot(
            permission_list=[robot_permission],
            duration=300,  # 5 minutes
            robot_name="test-system-robot-creator",
            robot_desc="System robot for testing project robot creation",
        )
        print("Created system robot: {}".format(system_robot.name))

        # System robot client configuration
        SYSTEM_ROBOT_CLIENT = dict(
            endpoint=TestRobotAccount.url,
            username=system_robot.name,
            password=system_robot.secret,
        )

        # Step 3: Use system robot to create a project-level robot
        try:
            project_robot_id, project_robot = self.robot.create_project_robot(
                project_name=project_name,
                duration=300,  # 5 minutes
                robot_name="test-project-robot-by-system",
                robot_desc="Project robot created by system robot",
                has_pull_right=True,
                has_push_right=True,
                **SYSTEM_ROBOT_CLIENT
            )
            print(
                "SUCCESS: System robot created project robot: {}".format(
                    project_robot.name
                )
            )

            # Step 4: Verify the project robot was created and has correct properties
            retrieved_robot = self.robot.get_robot_account_by_id(
                project_robot_id, **ADMIN_CLIENT
            )
            _assert_status_code(
                "project",
                retrieved_robot.level,
                "Expected level 'project', got '{}'".format(retrieved_robot.level),
            )
            _assert_status_code(
                1,
                len(retrieved_robot.permissions),
                "Expected 1 permission, got {}".format(
                    len(retrieved_robot.permissions)
                ),
            )
            _assert_status_code(
                project_name,
                retrieved_robot.permissions[0].namespace,
                "Expected namespace '{}', got '{}'".format(
                    project_name, retrieved_robot.permissions[0].namespace
                ),
            )

            print("SUCCESS: Project robot verification passed")

        except Exception as e:
            print(
                "FAILED: System robot could not create project robot: {}".format(str(e))
            )
            raise e

        finally:
            # Step 5: Clean up
            try:
                if "project_robot_id" in locals():
                    self.robot.delete_robot_account(project_robot_id, **ADMIN_CLIENT)
                    print("Cleaned up project robot: {}".format(project_robot_id))

                self.robot.delete_robot_account(system_robot_id, **ADMIN_CLIENT)
                print("Cleaned up system robot: {}".format(system_robot_id))

                self.project.delete_project(project_id, **ADMIN_CLIENT)
                print("Cleaned up project: {}".format(project_name))

            except Exception as cleanup_error:
                print("Warning: Cleanup error: {}".format(cleanup_error))


if __name__ == "__main__":
    suite = unittest.TestSuite(unittest.makeSuite(TestRobotAccount))
    result = unittest.TextTestRunner(sys.stdout, verbosity=2, failfast=True).run(suite)
    if not result.wasSuccessful():
        raise Exception(r"Robot account test failed: {}".format(result))
