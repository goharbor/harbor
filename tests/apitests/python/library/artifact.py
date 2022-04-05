# -*- coding: utf-8 -*-

import time
import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException


report_mime_types = [
    'application/vnd.security.vulnerability.report; version=1.1',
    'application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0',
]

class Artifact(base.Base, object):
    def __init__(self):
        super(Artifact,self).__init__(api_type = "artifact")

    def list_artifacts(self, project_name, repo_name, **kwargs):
        params = {}
        if "with_accessory" in kwargs:
            params["with_accessory"] = kwargs["with_accessory"]
        return self._get_client(**kwargs).list_artifacts(project_name, repo_name, **params)

    def get_reference_info(self, project_name, repo_name, reference, expect_status_code = 200, ignore_not_found = False,**kwargs):
        params = {}
        if "with_signature" in kwargs:
            params["with_signature"] = kwargs["with_signature"]
        if "with_tag" in kwargs:
            params["with_tag"] = kwargs["with_tag"]
        if "with_scan_overview" in kwargs:
            params["with_scan_overview"] = kwargs["with_scan_overview"]
            params["x_accept_vulnerabilities"] = ",".join(report_mime_types)
        if "with_immutable_status" in kwargs:
            params["with_immutable_status"] = kwargs["with_immutable_status"]
        if "with_accessory" in kwargs:
            params["with_accessory"] = kwargs["with_accessory"]

        try:
            data, status_code, _ = self._get_client(**kwargs).get_artifact_with_http_info(project_name, repo_name, reference, **params)
            return data
        except ApiException as e:
            if e.status == 404 and ignore_not_found == True:
                return None
            else:
                raise Exception("Failed to get reference, {} {}".format(e.status, e.body))
        else:
            base._assert_status_code(expect_status_code, status_code)
            base._assert_status_code(200, status_code)
            return None

    def delete_artifact(self, project_name, repo_name, reference, expect_status_code = 200, expect_response_body = None, **kwargs):
        try:
             _, status_code, _ = self._get_client(**kwargs).delete_artifact_with_http_info(project_name, repo_name, reference)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        else:
            base._assert_status_code(expect_status_code, status_code)
            base._assert_status_code(200, status_code)

    def get_addition(self, project_name, repo_name, reference, addition, **kwargs):
        return self._get_client(**kwargs).get_addition_with_http_info(project_name, repo_name, reference, addition)

    def add_label_to_reference(self, project_name, repo_name, reference, label_id, expect_status_code = 200, **kwargs):
        label = v2_swagger_client.Label(id = label_id)
        try:
            body, status_code, _ = self._get_client(**kwargs).add_label_with_http_info(project_name, repo_name, reference, label)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
        else:
            base._assert_status_code(expect_status_code, status_code)
            base._assert_status_code(200, status_code)
            return body

    def copy_artifact(self, project_name, repo_name, _from, expect_status_code = 201, expect_response_body = None, **kwargs):
        try:
            data, status_code, _ = self._get_client(**kwargs).copy_artifact_with_http_info(project_name, repo_name, _from)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        else:
            base._assert_status_code(expect_status_code, status_code)
            base._assert_status_code(201, status_code)
            return data

    def create_tag(self, project_name, repo_name, reference, tag_name, expect_status_code = 201, ignore_conflict = False, **kwargs):
        tag = v2_swagger_client.Tag(name = tag_name)
        try:
            _, status_code, _ = self._get_client(**kwargs).create_tag_with_http_info(project_name, repo_name, reference, tag)
        except ApiException as e:
            if e.status == 409 and ignore_conflict == True:
                return
            base._assert_status_code(expect_status_code, e.status)

        else:
            base._assert_status_code(expect_status_code, status_code)
            base._assert_status_code(201, status_code)

    def delete_tag(self, project_name, repo_name, reference, tag_name, expect_status_code = 200, **kwargs):
        try:
            _, status_code, _ = self._get_client(**kwargs).delete_tag_with_http_info(project_name, repo_name, reference, tag_name)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
        else:
            base._assert_status_code(expect_status_code, status_code)
            base._assert_status_code(200, status_code)

    def list_accessories(self, project_name, repo_name, reference, **kwargs):
        return self._get_client(**kwargs).list_accessories(project_name, repo_name, reference)

    def check_image_scan_result(self, project_name, repo_name, reference, expected_scan_status = "Success", **kwargs):
        timeout_count = 30
        scan_status=""
        while True:
            time.sleep(5)
            timeout_count = timeout_count - 1
            if (timeout_count == 0):
                break
            artifact = self.get_reference_info(project_name, repo_name, reference, **kwargs)
            if expected_scan_status in ["Not Scanned", "No Scan Overview"]:
                if artifact.scan_overview is None:
                    if (timeout_count > 24):
                        continue
                    print("artifact is not scanned.")
                    return
                else:
                    raise Exception("Artifact should not be scanned {}.".format(artifact.scan_overview))

            scan_status = ''
            for mime_type in report_mime_types:
                overview = artifact.scan_overview.get(mime_type)
                if overview:
                    scan_status = overview.scan_status

            if scan_status == expected_scan_status:
                return
        raise Exception("Scan image result is {}, not as expected {}.".format(scan_status, expected_scan_status))

    def check_reference_exist(self, project_name, repo_name, reference, ignore_not_found = False, **kwargs):
        artifact = self.get_reference_info( project_name, repo_name, reference, ignore_not_found=ignore_not_found, **kwargs)
        return {
            None: False,
        }.get(artifact, True)

    def waiting_for_reference_exist(self, project_name, repo_name, reference, ignore_not_found = True, period = 60, loop_count = 20, **kwargs):
        _loop_count = loop_count
        while True:
            print("Waiting for reference {} round...".format(_loop_count))
            _loop_count = _loop_count - 1
            if (_loop_count == 0):
                break
            artifact = self.get_reference_info(project_name, repo_name, reference, ignore_not_found=ignore_not_found, **kwargs)
            print("Returned artifact by get reference info:", artifact)
            if artifact  and artifact !=[]:
                return  artifact
            time.sleep(period)
        raise Exception("Reference is not exist {} {} {}.".format(project_name, repo_name, reference))
