from harborclient import base


class JobManager(base.Manager):
    def list(self, policy_id=None):
        """List filters jobs according to the policy and repository."""
        return self._list("/jobs/replication?policy_id=%s" % policy_id)

    def get_log(self, job_id):
        """Get job logs."""
        return self._get("/jobs/replication/%s/log" % job_id)
