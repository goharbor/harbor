---
title: Repositories
weight: 75
---

A repository is a collection of artifacts.  Since version v2.0, in addition to container images, Harbor can manage different kinds of artifacts that are bundled in an OCI-compatible format, such as a Helm chart (requires Helm v3), CNAB, or OPA bundle.

### List Repositories

Log in and click your project to view the project detail page. On the project detail page, click the **Repositories** tab to view the list of repositories. 

![list_repositories](../../../img/list-repositories.png)

### Description of a Repository

Click the repository, and then click the **Info** tab to see a description of the repository.  If you have the project admin, master, or developer role, you can click **Edit** to edit the description, which supports Markdown syntax.

![edit_repository_description](../../../img/edit-repository-description.png)

### List Artifacts in a Repository

To view the list of artifacts in a repository, click the **Artifacts** tab. 

Each artifact is identified by its `sha256` digest in the list of artifacts, and different types of artifacts are identified by the icon on the left of the digest. Hover your mouse over the icon to see the name of the type.  

To copy the command to pull the artifact, click the icon in the **Pull Command** column. The **Annotations** column shows the manifest annotations of the artifact, which are a set of key-value pairs.  For details about annotations, see (https://github.com/opencontainers/image-spec/blob/master/annotations.md). The **Push Time** column shows the time the artifact was pushed to the registry.

![list_artifacts](../../../img/list-artifacts.png)

Click the search icon in the top right to filter the list of artifacts. You can use different types of filters to filter the items in the artifact list. You can choose to filter by type, tags, labels. Particularly, if you choose to filter by tags, you can choose to view only the tagged or untagged artifacts.
<!--  this image is missing 
![filter_artifacts](../../../img/filter-artifacts.png)
  -->
  
Since Harbor v2.0, [Image index](https://github.com/opencontainers/image-spec/blob/master/image-index.md) can also be managed as an artifact in a repository.  If an artifact is an index, there is a folder icon on the right side of its digest.

![image_index](../../../img/index-icon.png)

When you click the folder icon, you can see a list of the artifacts referenced by the index.  Note that the artifacts in this view are read-only.  You can not remove an artifact from an index through Harbor's UI, and actions like 'copy digest', 'add labels', and 'copy' are not available.

![index_detail](../../../img/index-detail.png)
