# -*- coding: utf-8 -*-

import sys
import time
import base
import swagger_client

class Replication(base.Base):
    def create_replication_rule(self, 
        projectIDList, targetIDList, name=base._random_name("rule"), desc="", 
        filters=[], trigger=swagger_client.RepTrigger(kind="Manual"), 
        replicate_deletion=True,
        replicate_existing_image_now=False,
        **kwargs):
        projects = []
        for projectID in projectIDList:
            projects.append(swagger_client.Project(project_id=int(projectID)))
        targets = []
        for targetID in targetIDList:
            targets.append(swagger_client.RepTarget(id=int(targetID)))
        for filter in filters:
            filter["value"] = int(filter["value"])
        client = self._get_client(**kwargs)
        policy = swagger_client.RepPolicy(name=name, description=desc, 
            projects=projects, targets=targets, filters=filters, 
            trigger=trigger, replicate_deletion=replicate_deletion,
            replicate_existing_image_now=replicate_existing_image_now)
        _, _, header = client.policies_replication_post_with_http_info(policy)
        return base._get_id_from_header(header), name

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
