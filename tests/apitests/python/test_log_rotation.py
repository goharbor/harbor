from __future__ import absolute_import
import json
import time

import unittest

from testutils import ADMIN_CLIENT, suppress_urllib3_warning
from library.purge import Purge
from library.user import User


class TestLogRotation(unittest.TestCase, object):

    @suppress_urllib3_warning
    def setUp(self):
        self.purge = Purge()
        self.user = User()

    def tearDown(self):
        # 1. Reset schedule
        self.purge.update_purge_schedule(type=None, cron="", audit_retention_hour=0)

    def testLogRotation(self):
        """
        Test case:
            Log Rotaion API
        Test step and expected result:
            1. Create a purge audit log job;
            2. Stop this purge audit log job;
            3. Verify purge audit log job status is Stopped;
            4. Create a purge audit log job;
            5. Verify purge audit log job status is Success;
            6. Verify the log of the purge audit log job;
            7. Create purge audit log schedule;
            8. Verify purge audit log schedule;
            9. Update purge audit log schedule;
            10. Verify purge audit log schedule.
        Tear down:
            1 Reset schedule.
        """
        # 1. Create a purge audit log job
        self.purge.create_purge_schedule(type="Manual", cron=None, dry_run=True)
        # 2. Stop this purge audit log job
        latest_job = self.purge.get_latest_purge_job()
        self.purge.stop_purge_execution(latest_job.id)
        # 3. Verify purge audit log job status is Stopped
        # wait more 5s for status update after stop
        time.sleep(5)
        job_status = self.purge.get_purge_job(latest_job.id).job_status
        self.assertEqual(self.purge.get_purge_job(latest_job.id).job_status, "Stopped")
        # 4. Create a purge audit log job
        self.purge.create_purge_schedule(type="Manual", cron=None, dry_run=False, audit_retention_hour=1)
        # 5. Verify purge audit log job status is Success
        job_status = None
        job_id = None
        for i in range(20):
            print("wait for the job to finish:", i)
            if job_id == None:
                latest_job = self.purge.get_latest_purge_job()
                job_status = latest_job.job_status
                job_id = latest_job.id
            else:
                job_status = self.purge.get_purge_job(job_id).job_status
            if job_status == "Success":
                break
            time.sleep(2)
        self.assertEqual(job_status, "Success")
        # 6. Verify the log of the purge audit log job
        job_logs = self.purge.get_purge_job_log(job_id)
        self.assertIn("Purge audit job start", job_logs)
        self.assertIn("rows of audit logs", job_logs)
        # 7. Create a schedule
        schedule_type = "Weekly"
        schedule_cron = "0 0 0 * * 0"
        audit_retention_hour = 24
        include_event_types = "create_artifact,delete_artifact,pull_artifact"
        self.purge.create_purge_schedule(type=schedule_type, cron=schedule_cron, dry_run=False, audit_retention_hour=audit_retention_hour, include_event_types=include_event_types)
        # 8. Verify schedule
        self.verifySchedule(schedule_type, schedule_cron, audit_retention_hour, include_event_types)
        # 9. Update schedule
        schedule_type = "Custom"
        schedule_cron = "0 15 10 ? * *"
        audit_retention_hour = 12
        include_event_types = "create_artifact,delete_artifact"
        self.purge.update_purge_schedule(type=schedule_type, cron=schedule_cron, audit_retention_hour=audit_retention_hour, include_event_types=include_event_types)
        # 10. Verify schedule
        self.verifySchedule(schedule_type, schedule_cron, audit_retention_hour, include_event_types)

    def testLogRotationAPIPermission(self):
        """
        Test case:
            Log Rotaion Permission API
        Test step and expected result:
            1. Create a new user(UA);
            2. User(UA) should not have permission to create purge schedule API;
            3. Create a purge audit log job;
            4. User(UA) should not have permission to stop purge execution API;
            5. User(UA) should not have permission to get purge job API;
            6. User(UA) should not have permission to get purge job log API;
            7. User(UA) should not have permission to get purge jobs API;
            8. User(UA) should not have permission to get purge schedule API;
            9. User(UA) should not have permission to update purge schedule API;
        """
        expect_status_code = 403
        expect_response_body = "FORBIDDEN"
        # 1. Create a new user(UA)
        user_password = "Aa123456"
        _, user_name = self.user.create_user(user_password = user_password)
        USER_CLIENT = dict(endpoint = ADMIN_CLIENT["endpoint"], username = user_name, password = user_password)
        # 2. User(UA) should not have permission to create purge schedule API
        self.purge.create_purge_schedule(type="Manual", cron=None, dry_run=False, expect_status_code=expect_status_code, expect_response_body=expect_response_body, **USER_CLIENT)
        # 3. Create a purge audit log job
        self.purge.create_purge_schedule(type="Manual", cron=None, dry_run=False)
        latest_job = self.purge.get_latest_purge_job()
        # 4. User(UA) should not have permission to stop purge execution API
        self.purge.stop_purge_execution(latest_job.id, expect_status_code=expect_status_code, expect_response_body=expect_response_body, **USER_CLIENT)
        # 5. User(UA) should not have permission to get purge job API
        self.purge.get_purge_job(latest_job.id, expect_status_code=expect_status_code, expect_response_body=expect_response_body, **USER_CLIENT)
        # 6. User(UA) should not have permission to get purge job log API
        self.purge.get_purge_job_log(latest_job.id, expect_status_code=expect_status_code, expect_response_body=expect_response_body, **USER_CLIENT)
        # 7. User(UA) should not have permission to get purge jobs API
        self.purge.get_purge_jobs("creation_time", 10, 1, expect_status_code=expect_status_code, expect_response_body=expect_response_body, **USER_CLIENT)
        # 8. User(UA) should not have permission to get purge schedule API
        self.purge.get_purge_schedule(expect_status_code=expect_status_code, expect_response_body=expect_response_body, **USER_CLIENT)
        # 9. User(UA) should not have permission to update purge schedule API
        self.purge.update_purge_schedule(type="Custom", cron="0 15 10 ? * *", expect_status_code=expect_status_code, expect_response_body=expect_response_body, **USER_CLIENT)

    def verifySchedule(self, schedule_type, schedule_cron, audit_retention_hour, include_event_types):
        purge_schedule = self.purge.get_purge_schedule()
        job_parameters = json.loads(purge_schedule.job_parameters)
        self.assertEqual(purge_schedule.schedule.type, schedule_type)
        self.assertEqual(purge_schedule.schedule.cron, schedule_cron)
        self.assertEqual(job_parameters["audit_retention_hour"], audit_retention_hour)
        self.assertEqual(job_parameters["include_event_types"], include_event_types)

if __name__ == '__main__':
    unittest.main()