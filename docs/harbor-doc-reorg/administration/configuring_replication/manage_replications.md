# Manage Replications

You can list, add, edit and delete rules under `Administration->Replications`.   

![browse project](../img/manage_replication.png)

### Starting a replication manually
Select a replication rule and click `REPLICATE`, the resources which the rule is applied to will be replicated from the source registry to the destination immediately.  

![browse project](../img/start_replicate.png)

### Listing and stopping replication executions
Click a rule, the execution records which belong to this rule will be listed. Each record represents the summary of one execution of the rule. Click `STOP` to stop the executions which are in progress.  

![browse project](../img/list_stop_executions.png)

### Listing tasks
Click the ID of one execution, you can get the execution summary and the task list. Click the log icon can get the detail information for the replication progress.  
**Note**: The count of `IN PROGRESS` status in the summary includes both `Pending` and `In Progress` tasks.  

![browse project](../img/list_tasks.png)

### Deleting the replication rule
Select the replication rule and click `DELETE` to delete it. Only rules which have no in progress executions can be deleted.  

![browse project](../img/delete_rule.png)