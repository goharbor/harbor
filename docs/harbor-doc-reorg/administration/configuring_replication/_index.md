# Configuring Replication

Replication allows users to replicate resources (images/charts) between Harbor and non-Harbor registries in both pull or push mode. 

Once the system administrator has set a rule, all resources that match the defined [filter](#resource-filter) patterns will be replicated to the destination registry when the [triggering condition](#trigger-mode) is matched. Each resource will start a task to run. If the namespace does not exist on the destination registry, a new namespace will be created automatically. If it already exists and the user configured in the policy has no write privilege to it, the process will fail. The member information will not be replicated.  

There may be a bit of delay during replication based on the situation of the network. If a replication task fails, it will be re-scheduled a few minutes later and retried times.  

**Note:** Due to API changes, replication between different versions of Harbor is not supported.

- [Create Replication Endpoints](create_replication_endpoints.md)
- [Create Replication Rules](create_replication_rules.md)
- [Manage Replications](manage_replications.md)
