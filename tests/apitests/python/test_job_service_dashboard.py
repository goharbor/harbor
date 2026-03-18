from __future__ import absolute_import
import time

import unittest
import v2_swagger_client
from library import base
from testutils import harbor_server, ADMIN_CLIENT, suppress_urllib3_warning
from library.jobservice import Jobservice
from library.gc import GC
from library.purge import Purge
from library.user import User
from library.project import Project
from library.retention import Retention
from library.preheat import Preheat
from library.replication import Replication
from library.registry import Registry
from library.scan_all import ScanAll
from library.schedule import Schedule


class TestJobServiceDashboard(unittest.TestCase, object):

    @suppress_urllib3_warning
    def setUp(self):
        self.jobservice = Jobservice()
        self.gc = GC()
        self.purge = Purge()
        self.user = User()
        self.project = Project()
        self.retention = Retention()
        self.preheat = Preheat()
        self.replication = Replication()
        self.registry = Registry()
        self.scan_all = ScanAll()
        self.schedule = Schedule()
        self.job_types = [ "GARBAGE_COLLECTION", "PURGE_AUDIT_LOG", "P2P_PREHEAT", "IMAGE_SCAN", "REPLICATION", "RETENTION", "SCAN_DATA_EXPORT", "SCHEDULER", "SLACK", "SYSTEM_ARTIFACT_CLEANUP", "WEBHOOK", "EXECUTION_SWEEP", "AUDIT_LOGS_GDPR_COMPLIANT"]
        self.cron_type = "Custom"
        self.cron = "0 0 0 * * 0"

    def testJobQueues(self):
        """
        Test case:
            Job Service Dashboard Job Queues
        Test step and expected result:
            1. List job queue;
            2. Pause GC Job and purge audit Job;
            3. Verify that the Job status is Paused;
            4. Run GC and purge audit;
            5. Verify pending jobs of job queues;
            6. Resume GC Job and purge audit Job;
            7. Verify pending jobs of job queues;
            8. Pause GC Job and purge audit Job;
            9. Verify that the Job status is Paused;
            10. Run GC and purge audit;
            11. Verify pending jobs of job queues;
            12. Stop GC Job and purge audit Job;
            13. Verify pending jobs of job queues;
            14. Run GC and purge audit;
            15. Verify pending jobs of job queues;
            16. Stop all Job;
            17. Verify pending jobs of job queues;
            18. Resume all Job;
        """
        # 1. List job queue
        job_queues = self.jobservice.get_job_queues(**ADMIN_CLIENT)
        self.assertSetEqual(set(self.job_types), set(job_queues.keys()))

        # 2. Pause GC Job and purge audit Job
        self.jobservice.action_pending_jobs(self.job_types[0], "pause", **ADMIN_CLIENT)
        self.jobservice.action_pending_jobs(self.job_types[1], "pause", **ADMIN_CLIENT)

        # 3. Verify that the Job status is Paused
        job_queues = self.jobservice.get_job_queues(**ADMIN_CLIENT)
        self.assertTrue(job_queues[self.job_types[0]].paused)
        self.assertTrue(job_queues[self.job_types[1]].paused)

        # 4. Run GC and purge audit
        self.gc.gc_now(**ADMIN_CLIENT)
        self.purge.create_purge_schedule(type="Manual", cron=None, dry_run=False, **ADMIN_CLIENT)
        print(f"Start time: {time.time()}")
        time.sleep(2)
        print(f"End time: {time.time()}")

        # 5. Verify pending jobs of job queues
        print("Step 5 verifyPendingJobs\n")
        self.verifyPendingJobs([self.job_types[0], self.job_types[1]])

        # 6. Resume GC Job and purge audit Job
        self.jobservice.action_pending_jobs(self.job_types[0], "resume", **ADMIN_CLIENT)
        self.jobservice.action_pending_jobs(self.job_types[1], "resume", **ADMIN_CLIENT)

        # 7. Verify pending jobs of job queues
        self.waitJobQueuesStopToComplete([self.job_types[0], self.job_types[1]])

        # 8. Pause GC Job and purge audit Job;
        self.jobservice.action_pending_jobs(self.job_types[0], "pause", **ADMIN_CLIENT)
        self.jobservice.action_pending_jobs(self.job_types[1], "pause", **ADMIN_CLIENT)

        # 9. Verify that the Job status is Paused
        job_queues = self.jobservice.get_job_queues(**ADMIN_CLIENT)
        self.assertTrue(job_queues[self.job_types[0]].paused)
        self.assertTrue(job_queues[self.job_types[1]].paused)

        # 10. Run GC and purge audit
        self.gc.gc_now()
        self.purge.create_purge_schedule(type="Manual", cron=None, dry_run=False, **ADMIN_CLIENT)
        print(f"Start time: {time.time()}")
        time.sleep(2)
        print(f"End time: {time.time()}")
        
        # 11. Verify pending jobs of job queues
        print("Step 11 verifyPendingJobs\n")
        self.verifyPendingJobs([self.job_types[0], self.job_types[1]])

        # 12. Stop GC Job and purge audit Job
        self.jobservice.action_pending_jobs(self.job_types[0], "stop", **ADMIN_CLIENT)
        self.jobservice.action_pending_jobs(self.job_types[1], "stop", **ADMIN_CLIENT)

        # 13. Verify pending jobs of job queues
        self.waitJobQueuesStopToComplete([self.job_types[0], self.job_types[1]])

        # 14. Run GC and purge audit
        self.gc.gc_now(**ADMIN_CLIENT)
        self.purge.create_purge_schedule(type="Manual", cron=None, dry_run=False, **ADMIN_CLIENT)
        print(f"Start time: {time.time()}")
        time.sleep(2)
        print(f"End time: {time.time()}")

        # 15. Verify pending jobs of job queues
        print("Step 15 verifyPendingJobs\n")
        self.verifyPendingJobs([self.job_types[0], self.job_types[1]])

        # 16. Stop all Job
        self.jobservice.action_pending_jobs("all", "stop", **ADMIN_CLIENT)

        # 17. Verify pending jobs of job queues
        self.waitJobQueuesStopToComplete([self.job_types[0], self.job_types[1]])

        # 18. Resume all Job
        self.jobservice.action_pending_jobs("all", "resume", **ADMIN_CLIENT)
        job_queues = self.jobservice.get_job_queues(**ADMIN_CLIENT)
        self.assertFalse(job_queues[self.job_types[0]].paused)
        self.assertFalse(job_queues[self.job_types[1]].paused)


    def verifyPendingJobs(self, job_types):
        job_queues = self.jobservice.get_job_queues(**ADMIN_CLIENT)
        print("***** VerifyPendingJobs *****")
        for job_type in job_types:
            print(f"the count of job queue {job_type} is {job_queues[job_type].count}\n")
            print(f"the latency of job queue {job_type} is {job_queues[job_type].latency}\n")
            self.assertTrue(job_queues[job_type].count > 0)
            self.assertTrue(job_queues[job_type].latency > 0)
            self.assertTrue(job_queues[job_type].count > 0)
            self.assertTrue(job_queues[job_type].latency > 0)


    def waitJobQueuesStopToComplete(self, job_types):
        is_success = False
        for i in range(10):
            print("Wait for queues to be consumed:", i)
            job_queues = self.jobservice.get_job_queues()
            for job_type in job_types:
                if job_queues[job_type].count is not None or job_queues[job_type].latency is not None:
                    is_success = False
                    break
            else:
                is_success = True
                break
            time.sleep(2)
        self.assertTrue(is_success)


    def testSchedules(self):
        """
        Test case:
            Job Service Dashboard Schedules
        Test step and expected result:
            1. Create a new project;
            2. Create a retention policy triggered by schedule;
            3. Create a new distribution;
            4. Create a preheat policy triggered by schedule;
            5. Create a new registry;
            6. Create a replication policy triggered by schedule;
            7. Set up a schedule to scan all;
            8. Set up a schedule to GC;
            9. Set up a schedule to log rotation;
            10. Verify schedules;
            11. Pause all schedules;
            12. Verify schedules is Paused;
            13. Resume all schedules;
            14. Verify schedules is not Paused;
            15. Reset the schedule for scan all, GC, and log rotation;
            16. Verify schedules;
        """
        # 1. Create a new project(PA) by user(UA)
        project_id, project_name = self.project.create_project(metadata = {"public": "false"})

        # 2. Create a retention policy
        retention_id = self.retention.create_retention_policy(project_id, selector_repository="**", selector_tag="**")
        self.retention.update_retention_policy(retention_id, project_id, cron=self.cron)

        # 3. Create a new distribution
        _, distribution_name = self.preheat.create_instance(endpoint_url=base._random_name("https://"))

        # 4. Create a new preheat policy
        distribution = self.preheat.get_instance(distribution_name)
        _, preheat_policy_name = self.preheat.create_policy(project_name, project_id, distribution.id, trigger=r'{"type":"scheduled","trigger_setting":{"cron":"%s"}}' % (self.cron))
        preheat_policy = self.preheat.get_policy(project_name, preheat_policy_name)

        # 5. Create a new registry
        registry_id, _ = self.registry.create_registry("https://" + harbor_server)

        # 6. Create a replication policy triggered by schedule
        replication_id, _ = self.replication.create_replication_policy(dest_registry=v2_swagger_client.Registry(id=int(registry_id)), trigger=v2_swagger_client.ReplicationTrigger(type="scheduled",trigger_settings=v2_swagger_client.ReplicationTriggerSettings(cron=self.cron)))

        # 7. Set up a schedule to scan all
        self.scan_all.create_scan_all_schedule(self.cron_type, cron=self.cron)

        # 8. Set up a schedule to GC
        self.gc.create_gc_schedule(self.cron_type, is_delete_untagged=True, cron=self.cron)

        # 9. Set up a schedule to Log Rotation
        self.purge.create_purge_schedule(self.cron_type, self.cron, True)

        # 10. Verify schedules
        schedules = self.schedule.list_schedules(page_size=50, page=1)
        # 10.1 Verify retention schedule
        retention_schedule = schedules["%s-%d" % (self.job_types[5], retention_id)]
        self.assertEqual(retention_schedule.cron, self.cron)
        # 10.2 Verify preheat schedule
        preheat_schedule = schedules["%s-%d" % (self.job_types[2], preheat_policy.id)]
        self.assertEqual(preheat_schedule.cron, self.cron)
        # 10.3 Verify replication schedule
        replication_schedule = schedules["%s-%d" % (self.job_types[4], replication_id)]
        self.assertEqual(replication_schedule.vendor_type, self.job_types[4])
        self.assertEqual(replication_schedule.cron, self.cron)
        # 10.4 Verify scan all schedule
        scan_all_schedule = schedules["SCAN_ALL"]
        self.assertEqual(scan_all_schedule.cron, self.cron)
        # 10.5 Verify GC schedule
        gc_schedule = schedules[self.job_types[0]]
        self.assertEqual(gc_schedule.cron, self.cron)
        # 10.6 Verify log rotation
        log_rotation_schedule = schedules["PURGE_AUDIT_LOG"]
        self.assertEqual(log_rotation_schedule.cron, self.cron)

        # 11. Pause all schedules
        self.jobservice.action_pending_jobs("scheduler", "pause")

        # 12. Verify schedules is Paused;
        self.assertTrue(self.schedule.get_schedule_paused("all").paused)

        # 13. Resume all schedules
        self.jobservice.action_pending_jobs("scheduler", "resume")

        # 14. Verify schedules is not Paused
        self.assertFalse(self.schedule.get_schedule_paused("all").paused)

        # 15. Reset the schedule for scan all, GC, and log rotation
        self.scan_all.update_scan_all_schedule("None", cron="")
        self.gc.update_gc_schedule("None", cron="")
        self.purge.update_purge_schedule("None", "")

        # 16. Verify schedules
        schedules = self.schedule.list_schedules(page_size=50, page=1)
        # 16.1 Verify retention schedule
        retention_schedule = schedules["%s-%d" % (self.job_types[5], retention_id)]
        self.assertEqual(retention_schedule.cron, self.cron)
        # 16.2 Verify preheat schedule
        preheat_schedule = schedules["%s-%d" % (self.job_types[2], preheat_policy.id)]
        self.assertEqual(preheat_schedule.cron, self.cron)
        # 16.3 Verify replication schedule
        replication_schedule = schedules["%s-%d" % (self.job_types[4], replication_id)]
        self.assertEqual(replication_schedule.vendor_type, self.job_types[4])
        self.assertEqual(replication_schedule.cron, self.cron)
        # 16.4 Verify scan all schedule
        self.assertNotIn("SCAN_ALL", schedules)
        # 16.5 Verify GC schedule
        self.assertNotIn(self.job_types[0], schedules)
        # 16.6 Verify log rotation
        self.assertNotIn("PURGE_AUDIT_LOG", schedules)


    def testWorkers(self):
        """
        Test case:
            Job Service Dashboard Workers
        Test step and expected result:
            1. Get worker pools;
            2. Get workers in current pool;
            3. Stop running job;
        """
        # 1. Get worker pools
        worker_pools = self.jobservice.get_worker_pools()
        for worker_pool in worker_pools:
            self.assertIsNotNone(worker_pool.pid)
            self.assertIsNotNone(worker_pool.worker_pool_id)
            self.assertIsNotNone(worker_pool.concurrency)
            self.assertIsNotNone(worker_pool.start_at)
            self.assertIsNotNone(worker_pool.heartbeat_at)

            # 2. Get workers in current pool
            workers = self.jobservice.get_workers(worker_pool.worker_pool_id)
            self.assertEqual(len(workers), worker_pool.concurrency)
            for worker in workers:
                self.assertIsNotNone(worker.id)
                self.assertEqual(worker_pool.worker_pool_id, worker.pool_id)

        # 3. Stop running job
        self.jobservice.stop_running_job("966a49aa2278b67d743f8aca")


    def testJobServiceDashboardAPIPermission(self):
        """
        Test case:
            Log Rotaion Permission API
        Test step and expected result:
            1. Create a new user(UA);
            2. User(UA) should not have permission to list job queue API;
            3. User(UA) should not have permission to action_pending_jobs API;
            4. User(UA) should not have permission to stop running job API;
            5. User(UA) should not have permission to get worker pools API;
            6. User(UA) should not have permission to get workers in current pool API;
            7. User(UA) should not have permission to list schedules API;
            8. User(UA) should have permission to get scheduler paused status API;
            9. Verify that the get scheduler paused status API parameter job_type only support all;
        """
        expect_status_code = 403
        expect_response_body = "FORBIDDEN"

        # 1. Create a new user(UA)
        user_password = "Aa123456"
        _, user_name = self.user.create_user(user_password = user_password)
        USER_CLIENT = dict(endpoint = ADMIN_CLIENT["endpoint"], username = user_name, password = user_password)

        # 2. User(UA) should not have permission to list job queue API
        self.jobservice.get_job_queues(expect_status_code=expect_status_code, expect_response_body=expect_response_body, **USER_CLIENT)

        # 3. User(UA) should not have permission to action_pending_jobs API
        self.jobservice.action_pending_jobs(self.job_types[0], "pause", expect_status_code=expect_status_code, expect_response_body=expect_response_body, **USER_CLIENT)

        # 4. User(UA) should not have permission to stop running job API
        self.jobservice.stop_running_job("966a49aa2278b67d743f8aca", expect_status_code=expect_status_code, expect_response_body=expect_response_body, **USER_CLIENT)

        # 5. User(UA) should not have permission to get worker pools API
        self.jobservice.get_worker_pools(expect_status_code=expect_status_code, expect_response_body=expect_response_body, **USER_CLIENT)

        # 6. User(UA) should not have permission to get workers in current pool API
        self.jobservice.get_workers("13e0cfe999715102c47614ec", expect_status_code=expect_status_code, expect_response_body=expect_response_body, **USER_CLIENT)

        # 7. User(UA) should not have permission to list schedules API
        self.schedule.list_schedules(expect_status_code=expect_status_code, expect_response_body=expect_response_body, **USER_CLIENT)

        # 8. User(UA) should have permission to get scheduler paused status API
        self.schedule.get_schedule_paused("all")

        # 9. Verify that the get scheduler paused status API parameter job_type only support all
        self.schedule.get_schedule_paused(self.job_types[0], expect_status_code=400, expect_response_body="job_type can only be 'all'")


if __name__ == '__main__':
    unittest.main()
