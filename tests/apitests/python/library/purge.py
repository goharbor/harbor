# -*- coding: utf-8 -*-

import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException


class Purge(base.Base):

    def __init__(self):
        super(Purge, self).__init__(api_type="purge")

    def create_purge_schedule(self, type, cron, dry_run=True, audit_retention_hour=24, include_operations="create,delete,pull", expect_status_code=201, expect_response_body=None, **kwargs):
        scheduleObj = v2_swagger_client.ScheduleObj(type=type)
        if cron is not None:
            scheduleObj.cron = cron
        parameters = {
            "audit_retention_hour": audit_retention_hour,
            "include_operations": include_operations,
            "dry_run": dry_run
        }
        schedule = v2_swagger_client.Schedule(schedule=scheduleObj, parameters=parameters)
        try:
            _, status_code, _ = self._get_client(**kwargs).create_purge_schedule_with_http_info(schedule)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                print(e.body)
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)

    def update_purge_schedule(self, type, cron, audit_retention_hour=24, include_operations="create,delete,pull", expect_status_code=200, expect_response_body=None, **kwargs):
        scheduleObj = v2_swagger_client.ScheduleObj(type=type, cron=cron)
        parameters = {
            "audit_retention_hour": audit_retention_hour,
            "include_operations": include_operations,
            "dry_run": False
        }
        schedule = v2_swagger_client.Schedule(schedule=scheduleObj, parameters=parameters)
        try:
            _, status_code, _ = self._get_client(**kwargs).update_purge_schedule_with_http_info(schedule)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)

    def stop_purge_execution(self, purge_id, expect_status_code=200, expect_response_body=None, **kwargs):
        try:
            return_data, status_code, h = self._get_client(**kwargs).stop_purge_with_http_info(purge_id)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)

    def get_latest_purge_job(self, **kwargs):
        return self.get_purge_jobs(sort="-creation_time", page_size=1, page=1)[0]

    def get_purge_jobs(self, sort, page_size, page, expect_status_code=200, expect_response_body=None, **kwargs):
        try:
            return_data, status_code, _ = self._get_client(**kwargs).get_purge_history_with_http_info(sort=sort, page_size=page_size, page=page)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        return return_data

    def get_purge_job(self, purge_id, expect_status_code=200, expect_response_body=None, **kwargs):
        try:
            return_data, status_code, _ = self._get_client(**kwargs).get_purge_job_with_http_info(purge_id)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        return return_data

    def get_purge_job_log(self, purge_id, expect_status_code=200, expect_response_body=None, **kwargs):
        try:
            return_data, status_code, _ = self._get_client(**kwargs).get_purge_job_log_with_http_info(purge_id)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        return return_data

    def get_purge_schedule(self, expect_status_code=200, expect_response_body=None, **kwargs):
        try:
            return_data, status_code, _ = self._get_client(**kwargs).get_purge_schedule_with_http_info()
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        return return_data
