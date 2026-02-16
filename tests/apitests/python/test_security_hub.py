import unittest

from testutils import harbor_server, ADMIN_CLIENT, suppress_urllib3_warning
from library.project import Project
from library.user import User
from library.repository import push_image_to_project
from library.scan import Scan
from library.artifact import Artifact
from library.securityhub import SecurityHub

class TestSecurityHub(unittest.TestCase):


    @suppress_urllib3_warning
    def setUp(self):
        self.project = Project()
        self.user = User()
        self.scan = Scan()
        self.artifact = Artifact()
        self.securityhub = SecurityHub()
        self.image = "ghcr.io/goharbor/notary-server-photon"
        self.new_image = "notary-server-photon"
        self.tag = "v2.2.0"
        self.digest = "sha256:379bf2c7cd55b4214ced7f9a885b46f81992eb01abebfd67693f5cb394611ad1"


    def testSecurityHub(self):
        """
        Test case:
            Security Hub
        Test step and expected result:
            1. Create a new user(UA);
            2. Create a new project(PA) by user(UA);
            3. Push a new image(IA) in project(PA) by user(UA);
            4. Send scan image command and get tag(TA) information to check scan result, it should be finished;
            5. Get vulnerability system summary;
            6. Check the vulnerability system summary is correct;
            7. Check the Get the vulnerability list API;
            8. Check Get Vulnerability List API Search by severity;
            9. Check Get Vulnerability List API Search by project name;
            10. Check Get Vulnerability List API Search by repository name;
            11. Check Get Vulnerability List API Search by tag;
            12. Check Get Vulnerability List API Search by digest;
            13. Check Get Vulnerability List API Search by package;
            14. Check Get Vulnerability List API Search by cve_id;
            15. Check Get Vulnerability List API Search by cvss3;
            16. Check Get Vulnerability List API Search by all options;
        """
        url = ADMIN_CLIENT["endpoint"]
        user_password = "Aa123456"

        # 1. Create user(UA)
        _, user_name = self.user.create_user(user_password=user_password, **ADMIN_CLIENT)
        user_client = dict(endpoint=url, username=user_name, password=user_password, with_scan_overview = True)

        # 2. Create private project(PA) by user(UA)
        project_id, project_name = self.project.create_project(metadata={"public": "false"}, **user_client)

        # 3. Push a new image(IA) in project(PA) by user(UA)
        repository_name, tag = push_image_to_project(project_name, harbor_server, user_name, user_password, self.image, self.tag, new_image=self.new_image)

        # 4. Send scan image command and get tag(TA) information to check scan result, it should be finished
        self.scan.scan_artifact(project_name, self.new_image, tag, **user_client)
        self.artifact.check_image_scan_result(project_name, self.new_image, tag, **user_client)

        # 5. Get vulnerability system summary
        security_summary = self.securityhub.get_security_summary(**ADMIN_CLIENT)

        # 6. Check the vulnerability system summary is correct
        self.check_security_summary(security_summary, repository_name)

        # 7. Check the Get the vulnerability list API
        vulnerabilities = self.securityhub.list_vulnerabilities(**ADMIN_CLIENT)
        self.check_vulnerabilities(vulnerabilities)

        # 8. Check Get Vulnerability List API Search by severity
        vulnerabilities = self.securityhub.list_vulnerabilities(q="severity=Critical", **ADMIN_CLIENT)
        self.check_vulnerabilities(vulnerabilities, severity="Critical")
        vulnerabilities = self.securityhub.list_vulnerabilities(q="severity=High", **ADMIN_CLIENT)
        self.check_vulnerabilities(vulnerabilities, severity="High")
        vulnerabilities = self.securityhub.list_vulnerabilities(q="severity=Medium", **ADMIN_CLIENT)
        self.check_vulnerabilities(vulnerabilities, severity="Medium")
        vulnerabilities = self.securityhub.list_vulnerabilities(q="severity=Low", **ADMIN_CLIENT)
        self.check_vulnerabilities(vulnerabilities, severity="Low")

        # 9. Check Get Vulnerability List API Search by project name
        vulnerabilities = self.securityhub.list_vulnerabilities(q="project_id=%s" % project_id, **ADMIN_CLIENT)
        self.check_vulnerabilities(vulnerabilities, project_id=project_id)
        vulnerability = vulnerabilities[0]

        # 10. Check Get Vulnerability List API Search by repository name
        vulnerabilities = self.securityhub.list_vulnerabilities(q="repository_name=%s" % repository_name, **ADMIN_CLIENT)
        self.check_vulnerabilities(vulnerabilities, repository_name=repository_name)

        # 11. Check Get Vulnerability List API Search by tag
        vulnerabilities = self.securityhub.list_vulnerabilities(q="tag=%s" % tag, **ADMIN_CLIENT)
        self.check_vulnerabilities(vulnerabilities, tag=tag)

        # 12. Check Get Vulnerability List API Search by digest
        vulnerabilities = self.securityhub.list_vulnerabilities(q="digest=%s" % self.digest, **ADMIN_CLIENT)
        self.check_vulnerabilities(vulnerabilities, digest=self.digest)

        # 13. Check Get Vulnerability List API Search by package
        vulnerabilities = self.securityhub.list_vulnerabilities(q="package=%s" % vulnerability.package, **ADMIN_CLIENT)
        self.check_vulnerabilities(vulnerabilities, package=vulnerability.package)

        # 14. Check Get Vulnerability List API Search by cve_id
        vulnerabilities = self.securityhub.list_vulnerabilities(q="cve_id=%s" % vulnerability.cve_id, **ADMIN_CLIENT)
        self.check_vulnerabilities(vulnerabilities, cve_id=vulnerability.cve_id)

        # 15. Check Get Vulnerability List API Search by cvss3
        vulnerabilities = self.securityhub.list_vulnerabilities(q="cvss_score_v3=[%s~%s]" % (vulnerability.cvss_v3_score, vulnerability.cvss_v3_score), **ADMIN_CLIENT)
        self.check_vulnerabilities(vulnerabilities, cvss3_from=vulnerability.cvss_v3_score, cvss3_to=vulnerability.cvss_v3_score)

        # 16. Check Get Vulnerability List API Search by all options
        vulnerabilities = self.securityhub.list_vulnerabilities(q="severity=%s,project_id=%s,repository_name=%s,tag=%s,digest=%s,package=%s,cve_id=%s,cvss_score_v3=[%s~%s]" %
                                                                (vulnerability.severity, vulnerability.project_id, vulnerability.repository_name, vulnerability.tags[0], vulnerability.digest, vulnerability.package, vulnerability.cve_id, vulnerability.cvss_v3_score, vulnerability.cvss_v3_score), **ADMIN_CLIENT)
        self.check_vulnerabilities(vulnerabilities, severity=vulnerability.severity, project_id=vulnerability.project_id, repository_name=vulnerability.repository_name, tag=vulnerability.tags[0], digest=vulnerability.digest, package=vulnerability.package, cve_id=vulnerability.cve_id, cvss3_from=vulnerability.cvss_v3_score, cvss3_to=vulnerability.cvss_v3_score)

    def check_security_summary(self, security_summary, repository_name):
        # Check the summary is correct
        self.assertTrue(security_summary.critical_cnt > 0)
        self.assertTrue(security_summary.high_cnt > 0)
        self.assertTrue(security_summary.medium_cnt > 0)
        self.assertTrue(security_summary.low_cnt > 0)
        self.assertTrue(security_summary.fixable_cnt > 0)
        self.assertTrue(security_summary.scanned_cnt > 0)
        self.assertTrue(security_summary.total_artifact > 0)
        self.assertTrue(security_summary.total_vuls > 0)
        # Check the dangerous artifacts is correct
        dangerous_artifacts = security_summary.dangerous_artifacts
        self.assertTrue(0 < len(dangerous_artifacts) <= 5)
        artifact_sorted_list = sorted(dangerous_artifacts, key=lambda artifact: (artifact.critical_cnt, artifact.high_cnt, artifact.medium_cnt), reverse=True)
        self.assertListEqual(dangerous_artifacts, artifact_sorted_list)
        dangerous_artifacts_dict = {item.repository_name: item for item in dangerous_artifacts}
        self.assertIn(repository_name, dangerous_artifacts_dict.keys())
        for key, value in dangerous_artifacts_dict.items():
            if key == repository_name:
                self.assertEqual(value.digest, self.digest)
                self.assertTrue(value.critical_cnt > 0)
                self.assertTrue(value.high_cnt > 0)
                self.assertTrue(value.medium_cnt > 0)
                break
        # Check the dangerous cves is correct
        dangerous_cves = security_summary.dangerous_cves
        self.assertTrue(0 < len(dangerous_cves) <= 5)
        cve_sorted_list = sorted(dangerous_cves, key=lambda cve: float(cve.cvss_score_v3), reverse=True)
        self.assertListEqual(dangerous_cves, cve_sorted_list)


    def check_vulnerabilities(self, vulnerabilities, severity=None, cve_id=None, cvss3_from=0.0, cvss3_to=10.0, project_id=None, repository_name=None, package=None, tag=None, digest=None):
        for vulnerability in vulnerabilities:
            self.assertEqual(vulnerability.severity, severity) if severity is not None else self.assertIsNotNone(vulnerability.severity)
            self.assertEqual(vulnerability.cve_id, cve_id) if cve_id is not None else self.assertIsNotNone(vulnerability.cve_id)
            self.assertEqual(vulnerability.project_id, project_id) if project_id is not None else self.assertIsNotNone(vulnerability.project_id)
            self.assertEqual(vulnerability.repository_name, repository_name) if repository_name is not None else self.assertIsNotNone(vulnerability.repository_name)
            self.assertEqual(vulnerability.package, package) if package is not None else self.assertIsNotNone(vulnerability.package)
            self.assertEqual(vulnerability.tags[0], tag) if tag is not None else self.assertIsNotNone(vulnerability.tags[0])
            self.assertEqual(vulnerability.digest, digest) if digest is not None else self.assertIsNotNone(vulnerability.digest)
            self.assertTrue(cvss3_from <= vulnerability.cvss_v3_score <= cvss3_to)
            self.assertIsNotNone(vulnerability.desc)
            self.assertIsNotNone(vulnerability.links)
            self.assertIsNotNone(vulnerability.version)


    def testSecurityHubAPIValidate(self):
        """
        Test case:
            Log Rotaion Permission API
        Test step and expected result:
            1. Create a new user(UA);
            2. User(UA) should not have permission to call get vulnerability system summary API;
            3. User(UA) should not have permission to call get Vulnerability List API;
        """
        url = ADMIN_CLIENT["endpoint"]
        user_password = "Aa123456"
        # 1. Create user(UA)
        _, user_name = self.user.create_user(user_password=user_password, **ADMIN_CLIENT)
        user_client = dict(endpoint=url, username=user_name, password=user_password)
        # 2. User(UA) should not have permission to call get vulnerability system summary API
        self.securityhub.get_security_summary(expect_status_code=403, expect_response_body="FORBIDDEN", **user_client)
        # 3. User(UA) should not have permission to call get Vulnerability List API
        self.securityhub.list_vulnerabilities(expect_status_code=403, expect_response_body="FORBIDDEN", **user_client)


if __name__ == '__main__':
    unittest.main()
