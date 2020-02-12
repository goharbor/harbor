---
title: Creating a Replication Rule
weight: 25
---

A replication endpoint must exist before you create a replication rule. To create an endpoint, follow the instructions in [Creating Replication Endpoints](create-replication-endpoints.md).

1. Log in to the Harbor interface with an account that has Harbor system administrator privileges.
1. Expand **Administration**, and select **Replications**.

   ![Add a replication rule](../../../img/replication-rule1.png)
1. Click **New Replication Rule**.
1. Provide a name and description for the replication rule.
1. Select **Push-based** or **Pull-based** replication, depending on whether you want to replicate images to or from the remote registry.

   ![Replication mode](../../../img/replication-rule2.png)
1. For **Source resource filter**, identify the images to replicate.  

   ![Replication filters](../../../img/replication-rule3.png)

   * **Name**: Replicate resources with a given name by entering an image name or fragment.
   * **Tag**: Replicate resources with a given tag by entering a tag name or fragment.
   * **Label**: Replicate resources with a given label by using the drop-down menu to select from the available labels.
   * **Resource**: Replicate images, charts, or both.
   
   The name filter and tag filters support the following patterns:
   
   * **\***: Matches any sequence of non-separator characters `/`.
   * **\*\***: Matches any sequence of characters, including path separators `/`.
   * **?**: Matches any single non-separator character `/`.
   * **{alt1,...}**: Matches a sequence of characters if one of the comma-separated alternatives matches. are as follows:
   * **\***: Matches any sequence of non-separator characters `/`.
   * **\*\***: Matches any sequence of characters, including path separators `/`.
   * **?**: Matches any single non-separator character `/`.
   * **{alt1,...}**: Matches a sequence of characters if one of the comma-separated alternatives matches.
   
   **NOTE:** You must add `library` if you want to replicate the official images of Docker Hub. For example, `library/hello-world` matches the official hello-world images.  
   
   Pattern | String(Match or not)
   ---------- | -------
   `library/*`      | `library/hello-world`(Y)<br> `library/my/hello-world`(N)
   `library/**`     | `library/hello-world`(Y)<br> `library/my/hello-world`(Y)
   `{library,goharbor}/**` | `library/hello-world`(Y)<br> `goharbor/harbor-core`(Y)<br> `google/hello-world`(N)
   `1.?`      | `1.0`(Y)<br> `1.01`(N)
1. Use the **Destination Registry** drop-down menu to select from the configured replication endpoints. 
1. Enter the name of the namespace in which to replicate resources in the **Destination namespace** text box.

   If you do not enter a namespace, resources are placed in the same namespace as in the source registry. 

   ![Destination and namespaces](../../../img/replication-rule4.png)
1. Use the Trigger Mode drop-down menu to select how and when to run the rule.
   * **Manual**: Replicate the resources manually when needed. **Note**: Deletion operations are not replicated. 
   * **Scheduled**: Replicate the resources periodically by defining a cron job. **Note**: Deletion operations are not replicated. 
   * **Event Based**: When a new resource is pushed to the project, or an image is retagged, it is replicated to the remote registry immediately. If you select the **Delete remote resources when locally deleted**, if you delete an image, it is automatically deleted from the replication target.

  {{< note >}}
  You can filter images for replication based on the labels that are applied to the images. However, changing a label on an image does not trigger replication. Event-based replication is limited to pushing, retagging, and deleting images.
  {{< /note >}}

   ![Trigger mode](../../../img/replication-rule5.png)
      
1. Optionally select the Override checkbox to force replicated resources to replace resources at the destination with the same name.
1. Click **Save** to create the replication rule.  
