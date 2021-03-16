# -*- coding: utf-8 -*-

import time
import base
import v2_swagger_client
from v2_swagger_client.rest import ApiException

class ScanAll(base.Base):
    def __init__(self):
        super(ScanAll,self).__init__(api_type="scanall")

    def create_scan_all_schedule(self, schedule_type, cron=None, expect_status_code=201, expect_response_body=None, **kwargs):
        schedule_obj = v2_swagger_client.ScheduleObj()
        schedule_obj.type = schedule_type
        if cron is not None:
            schedule_obj.cron = cron

        schedule = v2_swagger_client.Schedule()
        schedule.schedule = schedule_obj

        try:
            _, status_code, _ = self._get_client(**kwargs).create_scan_all_schedule_with_http_info(schedule)
        except ApiException as e:
            if e.status == expect_status_code:
                if expect_response_body is not None and e.body.strip() != expect_response_body.strip():
                    raise Exception(r"Create scan all schedule response body is not as expected {} actual status is {}.".format(expect_response_body.strip(), e.body.strip()))
                else:
                    return e.reason, e.body
            else:
                raise Exception(r"Create scan all schedule result is not as expected {} actual status is {}.".format(expect_status_code, e.status))
        base._assert_status_code(expect_status_code, status_code)

    def scan_all_now(self, **kwargs):
        self.create_scan_all_schedule('Manual', **kwargs)

    def wait_until_scans_all_finish(self, **kwargs):
        client = self._get_client(**kwargs)
        timeout_count = 50
        while True:
            time.sleep(5)
            timeout_count = timeout_count - 1
            if (timeout_count == 0):
                break
            stats = client.get_latest_scan_all_metrics()
            print("Scan all status:", stats)
            if stats.ongoing is False:
                return
        raise Exception("Error: Scan all job is timeout.")
