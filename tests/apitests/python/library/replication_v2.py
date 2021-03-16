# -*- coding: utf-8 -*-

import time
import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException

class ReplicationV2(base.Base, object):
    def __init__(self):
        super(ReplicationV2,self).__init__(api_type = "replication")

    def wait_until_jobs_finish(self, rule_id, retry=10, interval=5, **kwargs):
        Succeed = False
        for i in range(retry):
            Succeed = False
            jobs = self.get_replication_executions(rule_id, **kwargs)
            for job in jobs:
                if job.status == "Succeed":
                    return
            if not Succeed:
                time.sleep(interval)
        if not Succeed:
            raise Exception("The jobs not Succeed")

    def trigger_replication_executions(self, rule_id, expect_status_code = 201, **kwargs):
        _, status_code, _ = self._get_client(**kwargs).start_replication_with_http_info({"policy_id":rule_id})
        base._assert_status_code(expect_status_code, status_code)

    def get_replication_executions(self, rule_id, expect_status_code = 200, **kwargs):
        data, status_code, _ = self._get_client(**kwargs).list_replication_executions_with_http_info(policy_id=rule_id)
        base._assert_status_code(expect_status_code, status_code)
        return data

