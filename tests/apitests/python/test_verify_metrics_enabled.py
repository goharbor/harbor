# coding: utf-8

from __future__ import absolute_import

import unittest
import requests
import testutils

class TestMetricsExist(unittest.TestCase):
    golang_basic_metrics = ["go_gc_duration_seconds", "go_goroutines", "go_info", "go_memstats_alloc_bytes"]

    metrics = {
        'core': golang_basic_metrics + [
            "harbor_core_http_request_total",
            "harbor_core_http_request_duration_seconds",
            "harbor_core_http_inflight_requests"],
        'registry': golang_basic_metrics + ["registry_http_in_flight_requests"],
        'exporter': golang_basic_metrics + [
            "artifact_pulled",
            "harbor_project_artifact_total",
            "harbor_project_member_total",
            "harbor_project_quota_byte",
            "harbor_project_repo_total",
            "harbor_project_total",
            "project_quota_usage_byte",
            "harbor_task_concurrency",
            "harbor_task_queue_latency",
            "harbor_task_queue_size",
            "harbor_task_scheduled_total"],
        'jobservice': golang_basic_metrics + [
            "harbor_jobservice_info",
            "harbor_jobservice_task_process_time_seconds",
            "harbor_jobservice_task_total"]
    }

    def get_metrics(self):
        metrics_url = testutils.METRIC_URL+'/metrics'
        exporter_res = requests.get(metrics_url)
        core_res = requests.get(metrics_url, params={'comp': 'core'})
        reg_res = requests.get(metrics_url, params={'comp': 'registry'})
        js_res = requests.get(metrics_url, params={'comp': 'jobservice'})
        return [('exporter', exporter_res.text), ('core', core_res.text), ('registry', reg_res.text), ('jobservice', js_res.text)]

    def testMetricsExist(self):
        for k, metric_text in self.get_metrics():



            for metric_name in self.metrics[k]:
                print("Metric {} should exist in {} ".format(metric_name, k))
                self.assertTrue(metric_name in metric_text)

if __name__ == '__main__':
    unittest.main()
