[Back to table of contents](../_index.md)

----------

# Configure Project Settings

You can configure Harbor so that only system administrators can create projects, and set limits on the amounts of resources that projects can consume.

## Set Who Can Create Projects

Use the **Project Creation** drop-down menu to set which users can create projects. Select **Everyone** to allow all users to create projects. Select **Admin Only** to allow only users with the Harbor system administrator role to create projects.  
![browse project](../../img/new_proj_create.png)

## Set Project Quotas

To exercise control over resource use, as a Harbor system administrator you can set  quotas on projects. You can limit the number of tags that a project can contain and limit the amount of storage capacity that a project can consume. You can set default quotas that apply to all projects globally.

**NOTE**: Default quotas apply to projects that are created after you set or change the default quota. The default quota is not applied to projects that already existed before you set it.

You can also set quotas on individual projects. If you set a global default quota and you set different quotas on individual projects, the per-project quotas are applied.

By default, all projects have unlimited quotas for both tags and storage use. 

1. Select the **Project Quotas** view.
   ![Project quotas](../../img/project-quota1.png)
1. To set global default quotas on all projects, click **Edit**.

   ![Project quotas](../../img/project-quota2.png)
   1. For **Default artifact count**, enter the maximum number of tags that any project can contain at a given time, or enter `-1` to set the default to unlimited.   
   1. For **Default storage consumption**, enter the maximum quantity of storage that any project can consume, selecting `MB`, `GB`, or `TB` from the drop-down menu, or enter `-1` to set the default to unlimited.  
   ![Project quotas](../../img/project-quota3.png)
   1. Click **OK**.
1. To set quotas on an individual project, click the 3 vertical dots next to a project and select **Edit**.
   ![Project quotas](../../img/project-quota4.png)
   1. For **Default artifact count**, enter the maximum number of tags that this individual project can contain, or enter `-1` to set the default to unlimited. 
   1. For **Default storage consumption**, enter the maximum quantity of storage that this individual project can consume, selecting `MB`, `GB`, or `TB` from the drop-down menu.

After you set quotas, you can see how much of their quotas each project has consumed.

![Project quotas](../../img/project-quota5.png)

### How Harbor Calculates Resource Usage

When setting project quotas, it is useful to know how Harbor calculates tag numbers and storage use, especially in relation to image pushing, retagging, and garbage collection.

- Harbor computes image size when blobs and manifests are pushed from the Docker client.
- Harbor computes tag counts when manifests are pushed from the Docker client.

   **NOTE**: When users push an image, the manifest is pushed last, after all of the associated blobs have been pushed successfully to the registry. If several images are pushed concurrently and if there is an insufficient number of tags left in the quota for all of them, images are accepted in the order that their manifests arrive. Consequently, an attempt to push an image might not be immediately rejected for exceeding the quota. This is because there was availability in the tag quota when the push was initiated, but by the time the manifest arrived the quota had been exhausted.
- Shared blobs are only computed once per project. In Docker, blob sharing is defined globally. In Harbor, blob sharing is defined at the project level. As a consequence, overall storage usage can be greater than the actual disk capacity.
- Retagging images reserves and releases resources: 
  -  If you retag an image within a project, the tag count increases by one, but storage usage does not change because there are no new blobs or manifests.
  - If you retag an image from one project to another, the tag count and storage usage both increase.
- During garbage collection, Harbor frees the storage used by untagged blobs in the project.
- If the tag count reaches the limit, image blobs can be pushed into a project and storage usage is updated accordingly. You can consider these blobs to be untagged blobs. They can be removed by garbage collection, and the storage that they consume is returned after garbage colletion.
- Helm chart size is not calculated. Only tag counts are calculated.

----------

[Back to table of contents](../_index.md)
