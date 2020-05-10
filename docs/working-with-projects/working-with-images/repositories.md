---
title: Repositories
weight: 75
---

A repository is a collection of artifacts.  Since version v2.0, in addition to container images Harbor can manage kinds 
artifacts that are bundled in OCI-compatible format.

### List Repositories

Click your project to enter the project detail page after successful logging in.  Click the "Repositories" tab to view the list of of repositories. 

![list_repositories](../../../img/list-repositories.png)

### Description of a repository

Click the repository, then click the "Info" tab.  You can view the description of the project.  Users with project admin, master or developer or master role can click the "Edit" button to edit the description.  You can style the description via Markdown syntax.

![edit_repository_description](../../../img/edit-repository-description.png)

### List artifacts in a repository

Click the "Artifacts" tab to view the list of artifacts in a repository.
Each artifact is identified by the sha256 digest in the list of artifacts, and different types of artifacts can be distinguished by the icon on the left of the digest.  Hover your mouse on the icon you can see the name of the type.  

By clicking the icon in the column **Pull Command**, the command to pull the artifact in the row of the icon will be copied to the clipboard.  
The column **Annotations** in the grid shows the manifest annotations of the artifact, which are a set of key-value pairs.  More details about the annotations please refer to (https://github.com/opencontainers/image-spec/blob/master/annotations.md).
The column **Push Time** in the grid shows the time each artifact is pushed to the registry.

![list_artifacts](../../../img/list-artifacts.png)

By clicking the search icon in the top right of the list of artifacts, you can different types of filter to filter the items in the artifact list.  You can choose to filter by type, tags, labels.  Particularly, if you choose to filter by tags, you choose to view only the tagged or untagged artifacts.

![filter_artifacts](../../../img/filter-artifacts.png)
