# -*- coding: utf-8 -*-

import base
from v2_swagger_client.rest import ApiException


class Schedule(base.Base):

    def __init__(self):
        super(Schedule, self).__init__(api_type="schedule")

    def get_schedule_paused(self, job_type, expect_status_code=200, expect_response_body=None, **kwargs):
        try:
            return_data, status_code, _ = self._get_client(**kwargs).get_schedule_paused_with_http_info(job_type)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        return return_data

    def list_schedules(self, page_size=50, page=1, expect_status_code=200, expect_response_body=None, **kwargs):
        try:
            schedules, status_code, _ = self._get_client(**kwargs).list_schedules_with_http_info(page_size=50, page=1)
        except ApiException as e:
            base._assert_status_code(expect_status_code, e.status)
            if expect_response_body is not None:
                base._assert_status_body(expect_response_body, e.body)
            return
        base._assert_status_code(expect_status_code, status_code)
        schedule_dict = {}
        for schedule in schedules:
            if schedule.vendor_id in [ None, -1]:
                schedule_dict[schedule.vendor_type] = schedule
            else:
                schedule_dict["%s-%d" % (schedule.vendor_type, schedule.vendor_id)] = schedule
        return schedule_dict
