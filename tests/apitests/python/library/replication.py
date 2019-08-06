# -*- coding: utf-8 -*-

import sys
import time
import base
import swagger_client

class Replication(base.Base):

    def get_replication_rule(self, param = None, rule_id = None, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        if rule_id is None:
            if param is None:
                param = dict()
            data, status_code, _ = client.policies_replication_get_with_http_info(param)
        else:
            data, status_code, _ = client.policies_replication_id_get_with_http_info(rule_id)
        base._assert_status_code(expect_status_code, status_code)
        return data

    def check_replication_rule_should_exist(self, check_rule_id, expect_rule_name, expect_trigger = None, **kwargs):
        rule_data = self.get_replication_rule(rule_id = check_rule_id, **kwargs)
        if str(rule_data.name) != str(expect_rule_name):
            raise Exception(r"Check replication rule failed, expect <{}> actual <{}>.".format(expect_rule_name, str(rule_data.name)))
        else:
            print r"Check Replication rule passed, rule name <{}>.".format(str(rule_data.name))
            get_trigger = str(rule_data.trigger.kind)
            if expect_trigger is not None and get_trigger == str(expect_trigger):
                print r"Check Replication rule trigger passed, trigger name <{}>.".format(get_trigger)
            else:
                raise Exception(r"Check replication rule trigger failed, expect <{}> actual <{}>.".format(expect_trigger, get_trigger))


    def start_replication(self, rule_id, **kwargs):
        client = self._get_client(**kwargs)
        return client.replications_post(swagger_client.Replication(int(rule_id)))

    def list_replication_jobs(self, rule_id, **kwargs):
        client = self._get_client(**kwargs)
        return client.jobs_replication_get(int(rule_id))

    def wait_until_jobs_finish(self, rule_id, retry=10, interval=5, **kwargs):
        finished = True
        for i in range(retry):
            finished = True
            jobs = self.list_replication_jobs(rule_id, **kwargs)
            for job in jobs:
                if job.status != "finished":
                    finished = False
                    break
            if not finished:
                time.sleep(interval)
        if not finished:
            raise Exception("The jobs not finished")

    def delete_replication_rule(self, rule_id, expect_status_code = 200, **kwargs):
        client = self._get_client(**kwargs)
        _, status_code, _ = client.policies_replication_id_delete_with_http_info(rule_id)
        base._assert_status_code(expect_status_code, status_code)
