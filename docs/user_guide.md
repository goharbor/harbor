# User Guide  
## Overview  
This guide walks you through the fundamentals of using Harbor. You'll learn how to use Harbor to:  

* [Manage your projects](#managing-projects)
* [Manage members of a project](#managing-members-of-a-project)
* [Replicate resources between Harbor and non-Harbor registries](#replicating-resources)
* [Retag images within Harbor](#retag-images)
* [Search projects and repositories](#searching-projects-and-repositories)
* [Manage labels](#managing-labels)
* [Set Project Quotas](#set-project-quotas)
* [Manage Harbor system if you are the system administrator:](#administrator-options)
  * [Manage users](#managing-user)
  * [Manage registries](#managing-registry)
  * [Manage replication rules](#managing-replication)
  * [Manage authentication](#managing-authentication)
  * [Manage project creation](#managing-project-creation)
  * [Manage self-registration](#managing-self-registration)
  * [Manage email settings](#managing-email-settings)
  * [Manage registry read only](#managing-registry-read-only)
  * [Manage role by LDAP group](#managing-role-by-ldap-group)
* [Pull and push images using Docker client](#pulling-and-pushing-images-using-docker-client)
* [Add description to repositories](#add-description-to-repositories)
* [Delete repositories and images](#deleting-repositories)
* [Content trust](#content-trust)
* [Vulnerability Scanning](#vulnerability-scanning)
* [Configure CVE Whitelists](#configure-cve-whitelists)
* [Pull image from Harbor in Kubernetes](#pull-image-from-harbor-in-kubernetes)
* [Manage Helm Charts](#manage-helm-charts)
  * [Manage Helm Charts via portal](#manage-helm-charts-via-portal)
  * [Working with Helm CLI](#working-with-helm-cli)
* [Garbage Collection](#garbage-collection)
* [View build history](#build-history)
* [Using CLI after login via OIDC based SSO](#using-oidc-cli-secret)
* [Manage robot account of a project](#robot-account)
* [Tag Retention Rules](#tag-retention-rules)
* [Configure Tag Immutability](#configure-tag-immutability)
* [Webhook Notifications](#webhook-notifications)
* [Using API Explorer](#api-explorer)

## Role Based Access Control(RBAC)  

![rbac](img/rbac.png)

Harbor manages images through projects. Users can be added into one project as a member with one of the following different roles:  

* **Limited Guest**: A Limited Guest does not have full read privileges for a project. They can pull images but cannot push, and they cannot see logs or the other members of a project. For example, you can create limited guests for users from different organizations who share access to a project.
* **Guest**: Guest has read-only privilege for a specified project. They can pull and retag images, but cannot push.
* **Developer**: Developer has read and write privileges for a project.
* **Master**: Master has elevated permissions beyond those of 'Developer' including the ability to scan images, view replications jobs, and delete images and helm charts. 
* **ProjectAdmin**: When creating a new project, you will be assigned the "ProjectAdmin" role to the project. Besides read-write privileges, the "ProjectAdmin" also has some management privileges, such as adding and removing members, starting a vulnerability scan.

Besides the above roles, there are two system-level roles:  

* **SysAdmin**: "SysAdmin" has the most privileges. In addition to the privileges mentioned above, "SysAdmin" can also list all projects, set an ordinary user as administrator, delete users and set vulnerability scan policy for all images. The public project "library" is also owned by the administrator.  
* **Anonymous**: When a user is not logged in, the user is considered as an "Anonymous" user. An anonymous user has no access to private projects and has read-only access to public projects.  

See detailed permissions matrix listed here: https://github.com/goharbor/harbor/blob/master/docs/permissions.md

## User account
Harbor supports different authentication modes:  

* **Database(db_auth)**  

	Users are stored in the local database.  
	
	A user can register himself/herself in Harbor in this mode. To disable user self-registration, refer to the [installation guide](installation_guide.md) for initial configuration, or disable this feature in [Administrator Options](#administrator-options). When self-registration is disabled, the system administrator can add users into Harbor.  
	
	When registering or adding a new user, the username and email must be unique in the Harbor system. The password must contain at least 8 characters with 1 lowercase letter, 1 uppercase letter and 1 numeric character.
	
	When you forgot your password, you can follow the below steps to reset the password:  

	1. Click the link "Forgot Password" in the sign in page.  
	2. Input the email address entered when you signed up, an email will be sent out to you for password reset.  
	3. After receiving the email, click on the link in the email which directs you to a password reset web page.  
	4. Input your new password and click "Save".  
	
* **LDAP/Active Directory (ldap_auth)**  

	Under this authentication mode, users whose credentials are stored in an external LDAP or AD server can log in to Harbor directly.  
	
	When an LDAP/AD user logs in by *username* and *password*, Harbor binds to the LDAP/AD server with the **"LDAP Search DN"** and **"LDAP Search Password"** described in [installation guide](installation_guide.md). If it succeeded, Harbor looks up the user under the LDAP entry **"LDAP Base DN"** including substree. The attribute (such as uid, cn) specified by **"LDAP UID"** is used to match a user with the *username*. If a match is found, the user's *password* is verified by a bind request to the LDAP/AD server. Uncheck **"LDAP Verify Cert"** if the LDAP/AD server uses a self-signed or an untrusted certificate.
	
	Self-registration, deleting user, changing password and resetting password are not supported under LDAP/AD authentication mode because the users are managed by LDAP or AD.  

* **OIDC Provider (oidc_auth)**

    With this authentication mode, regular user will login to Harbor Portal via SSO flow.  
    After the system administrator configure Harbor to authenticate via OIDC (more details refer to [this section](#managing-authentication)),
    a button `LOGIN VIA OIDC PROVIDER` will appear on the login page.  
    ![oidc_login](img/oidc_login.png)
    
    By clicking this button user will kick off the SSO flow and be redirected to the OIDC Provider for authentication.  After a successful
    authentication at the remote site, user will be redirected to Harbor.  There will be an "onboard" step if it's the first time the user 
    authenticate using his account, in which there will be a dialog popped up for him to set his user name in Harbor:
    ![oidc_onboar](img/oidc_onboard_dlg.png)
    
    This user name will be the identifier for this user in Harbor, which will be used in the cases such as adding member to a project, assigning roles, etc.
    This has to be a unique user name, if another user has used this user name to onboard, user will be prompted to choose another one.
    
    Regarding this user to use docker CLI, please refer to [Using CLI after login via OIDC based SSO](#using-oidc-cli-secret)
   
    **NOTE:**
    1. After the onboard process, you still have to login to Harbor via SSO flow, the `Username` and `Password` fields are only for
    local admin to login when Harbor is configured authentication via OIDC.
    2. Similar to LDAP authentication mode, self-registration, updating profile, deleting user, changing password and 
    resetting password are not supported.
 

## Managing projects
A project in Harbor contains all repositories of an application. No images can be pushed to Harbor before the project is created. RBAC is applied to a project. There are two types of projects in Harbor:  

* **Public**: All users have the read privilege to a public project, it's convenient for you to share some repositories with others in this way.
* **Private**: A private project can only be accessed by users with proper privileges.  

You can create a project after you signed in. Check on the "Access Level" checkbox will make this project public.  

![create project](img/new_create_project.png)  

After the project is created, you can browse repositories, members, logs, replication and configuration using the navigation tab.

![browse project](img/new_browse_project.png)

There are two views to show repositories, list view and card view, you can switch between them by clicking the corresponding icon.

![browse repositories](img/browse_project_repositories.png)

All logs can be listed by clicking "Logs". You can apply a filter by username, or operations and dates under "Advanced Search".  

![browse project](img/log_search_advanced.png)

![browse project](img/new_project_log.png)  

Project properties can be changed by clicking "Configuration".

* To make all repositories under the project accessible to everyone, select the `Public` checkbox.

* To prevent un-signed images under the project from being pulled, select the `Enable content trust` checkbox.

* To prevent vulnerable images under the project from being pulled, select the `Prevent vulnerable images from running` checkbox and change the severity level of vulnerabilities. Images cannot be pulled if their level equals to or higher than the currently selected level.

* To activate an immediate vulnerability scan on new images that are pushed to the project, select the `Automatically scan images on push` checkbox.

![browse project](img/project_configuration.png) 

## Managing members of a project  
### Adding members  
You can add members with different roles to an existing project. You can add an LDAP/AD user to project members under LDAP/AD authentication mode. 

![browse project](img/new_add_member.png)

### Updating and removing members
You can check one or more members, then click `ACTION`, choose one role to batch switch checked members' roles or remove them from the project.

![browse project](img/new_remove_update_member.png)

## Replicating resources  
Replication allows users to replicate resources (images/charts) between Harbor and non-Harbor registries in both pull or push mode. 

Once the system administrator has set a rule, all resources that match the defined [filter](#resource-filter) patterns will be replicated to the destination registry when the [triggering condition](#trigger-mode) is matched. Each resource will start a task to run. If the namespace does not exist on the destination registry, a new namespace will be created automatically. If it already exists and the user configured in the policy has no write privilege to it, the process will fail. The member information will not be replicated.  

There may be a bit of delay during replication based on the situation of the network. If a replication task fails, it will be re-scheduled a few minutes later and retried times.  

**Note:** Due to API changes, replication between different versions of Harbor is not supported.

### Creating replication endpoints

To replicate image repositories from one instance of Harbor to another Harbor or non-Harbor registry, you first create replication endpoints.

1. Go to **Registries** and click the **+ New Endpoint** button.

   ![New replication endpoint](img/replication-endpoint1.png)
1. For **Provider**, use the drop-down menu to select the type of registry to set up as a replication endpoint.

   The endpoint can be another Harbor instance, or a non-Harbor registry. Currently, the following non-Harbor registries are supported:

   - Docker Hub
   - Docker registry
   - AWS Elastic Container Registry
   - Azure Container Registry
   - Ali Cloud Container Registry
   - Google Container Registry
   - Huawei SWR
   - Helm Hub
   - Gitlab
   - Quay.io
   - Jfrog Artifactory

   ![Replication providers](img/replication-endpoint2.png)

1. Enter a suitable name and description for the new replication endpoint.
1. Enter the full URL of the registry to set up as a replication endpoint.

   For example, to replicate to another Harbor instance, enter https://harbor_instance_address:443. The registry must exist and be running before you create the endpoint.
1. Enter the Access ID and Access Secret for the endpoint registry instance.

   Use an account that has the appropriate privileges on that registry, or an account that has write permission on the corresponding project in a Harbor  registry.
   
   **NOTES**: 
    - AWS ECR adapters should use access keys, not a username and password. The access key should have sufficient permissions, such as storage permission.
    - Google GCR adapters should use the entire JSON key generated in the service account. The namespace should start with the project ID.
1. Optionally, select the **Verify Remote Cert** check box.

    Deselect the check box if the remote registry uses a self-signed or untrusted certificate.
1. Click **Test Connection**.
1. When you have successfully tested the connection, click **OK**.

### Creating a replication rule
Login as a system administrator user, click `NEW REPLICATION RULE` under `Administration->Replications` and fill in the necessary fields. You can choose different replication modes, [resource filters](#resource-filter) and [trigger modes](#trigger-mode) according to the different requirements. If there is no endpoint available in the list, follow the instructions in the [Creating replication endpoints](#creating-replication-endpoints) to create one. Click `SAVE` to create a replication rule.  

![browse project](img/create_rule.png)

#### Resource filter
Three resource filters are supported:
* **Name**: Filter resources according to the name.
* **Tag**: Filter resources according to the tag.
* **Resource**: Filter images according to the resource type.

The terms supported in the pattern used by name filter and tag filter are as follows:
* **\***: Matches any sequence of non-separator characters `/`.
* **\*\***: Matches any sequence of characters, including path separators `/`.
* **?**: Matches any single non-separator character `/`.
* **{alt1,...}**: Matches a sequence of characters if one of the comma-separated alternatives matches.

**Note:** `library` must be added if you want to replicate the official images of Docker Hub. For example, `library/hello-world` matches the official hello-world images.  

Pattern | String(Match or not)
---------- | -------
`library/*`      | `library/hello-world`(Y)<br> `library/my/hello-world`(N)
`library/**`     | `library/hello-world`(Y)<br> `library/my/hello-world`(Y)
`{library,goharbor}/**` | `library/hello-world`(Y)<br> `goharbor/harbor-core`(Y)<br> `google/hello-world`(N)
`1.?`      | `1.0`(Y)<br> `1.01`(N)

#### Trigger mode
* **Manual**: Replicate the resources manually when needed. **Note**: The deletion operations are not replicated. 
* **Scheduled**: Replicate the resources periodically. **Note**: The deletion operations are not replicated. 
* **Event Based**: When a new resource is pushed to the project, it is replicated to the remote registry immediately. Same to the deletion operation if the `Delete remote resources when locally deleted` checkbox is selected.

### Starting a replication manually
Select a replication rule and click `REPLICATE`, the resources which the rule is applied to will be replicated from the source registry to the destination immediately.  

![browse project](img/start_replicate.png)

### Listing and stopping replication executions
Click a rule, the execution records which belong to this rule will be listed. Each record represents the summary of one execution of the rule. Click `STOP` to stop the executions which are in progress.  

![browse project](img/list_stop_executions.png)

### Listing tasks
Click the ID of one execution, you can get the execution summary and the task list. Click the log icon can get the detail information for the replication progress.  
**Note**: The count of `IN PROGRESS` status in the summary includes both `Pending` and `In Progress` tasks.  

![browse project](img/list_tasks.png)

### Deleting the replication rule
Select the replication rule and click `DELETE` to delete it. Only rules which have no in progress executions can be deleted.  

![browse project](img/delete_rule.png)

## Retag Images

Images retag helps users to tag images in Harbor, images can be tagged to  different repositories and projects, as long as the users have sufficient permissions. For example,

```
release/app:stg  -->  release/app:prd
develop/app:v1.0 --> release/app:v1.0
```
To retag an image, users should have read permission (guest role or above) to the source project and write permission (developer role or above) to the target project.

In Harbor portal, select the image you'd like to retag, and click the enabled `Retag` button to open the retag dialog.

![retag image](img/retag_image.png)

In the retag dialog, project name, repository name and the new tag should be specified. On click the `CONFIRM` button, the new tag would be created instantly. You can check the new tag in the corresponding project. 

## Searching projects and repositories
Entering a keyword in the search field at the top lists all matching projects and repositories. The search result includes both public and private repositories you have access to.  

![browse project](img/new_search.png)

## Managing labels
Harbor provides two kinds of labels to isolate kinds of resources(only images for now):
* **Global Level Label**: Managed by system administrators and used to manage the images of the whole system. They can be added to images under any projects.
* **Project Level Label**: Managed by project administrators under a project and can only be added to the images of the project.

### Managing global level labels
The system administrators can list, create, update and delete the global level labels under `Administration->Configuration->Labels`:

![manage global level labels](img/manage_global_level_labels.png)

### Managing project level labels
The project administrators and system administrators can list, create, update and delete the project level labels under `Labels` tab of the project detail page:

![manage project level labels](img/manage_project_level_labels.png)

### Adding labels to/remove labels from images
Users who have system administrator, project administrator or project developer role can click the `ADD LABELS` button to add labels to or remove labels from images. The label list contains both globel level labels(come first) and project level labels:

![add labels to images](img/add_labels_to_images.png)

### Filtering images by labels
The images can be filtered by labels:

![filter images by labels](img/filter_images_by_label.png)

## Set Project Quotas

To exercise control over resource use, as a Harbor administrator you can set  quotas on projects. You can limit the number of tags that a project can contain and limit the amount of storage capacity that a project can consume. You can set default quotas that apply to all projects globally.

**NOTE**: Default quotas apply to projects that are created after you set or change the default quota. The default quota is not applied to projects that already existed before you set it.

You can also set quotas on individual projects. If you set a global default quota and you set different quotas on individual projects, the per-project quotas are applied.

By default, all projects have unlimited quotas for both tags and storage use. 

1. Select the **Project Quotas** view.
   ![Project quotas](img/project-quota1.png)
1. To set global default quotas on all projects, click **Edit**.

   ![Project quotas](img/project-quota2.png)
   1. For **Default artifact count**, enter the maximum number of tags that any project can contain at a given time, or enter `-1` to set the default to unlimited.   
   1. For **Default storage consumption**, enter the maximum quantity of storage that any project can consume, selecting `MB`, `GB`, or `TB` from the drop-down menu, or enter `-1` to set the default to unlimited.  
   ![Project quotas](img/project-quota3.png)
   1. Click **OK**.
1. To set quotas on an individual project, click the 3 vertical dots next to a project and select **Edit**.
   ![Project quotas](img/project-quota4.png)
   1. For **Default artifact count**, enter the maximum number of tags that this individual project can contain, or enter `-1` to set the default to unlimited. 
   1. For **Default storage consumption**, enter the maximum quantity of storage that this individual project can consume, selecting `MB`, `GB`, or `TB` from the drop-down menu.

After you set quotas, you can see how much of their quotas each project has consumed.

![Project quotas](img/project-quota5.png)

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

## Administrator options  
### Managing user  
Administrator can add "Administrator" role to one or more ordinary users by checking checkboxes and clicking `SET AS ADMINISTRATOR`. To delete users, checked checkboxes and select `DELETE`. Deleting user is only supported under database authentication mode.

![browse project](img/new_set_admin_remove_user.png)

### Managing registry  
You can list, add, edit and delete registries under `Administration->Registries`. Only registries which are not referenced by any rules can be deleted.  

![browse project](img/manage_registry.png)

### Managing replication  
You can list, add, edit and delete rules under `Administration->Replications`.   

![browse project](img/manage_replication.png)

### Managing authentication
You can change authentication mode between **Database**(default) and **LDAP** before any user is added, when there is at least one user(besides admin) in Harbor, you cannot change the authentication mode.  
![browse project](img/new_auth.png)
When using LDAP mode, user's self-registration is disabled. The parameters of LDAP server must be filled in. For more information, refer to [User account](#user-account).   
![browse project](img/ldap_auth.png)

When using OIDC mode, user will login Harbor via OIDC based SSO.  A client has to be registered on the OIDC provider and Harbor's callback URI needs to be associated to that client as a redirectURI.
![OIDC settings](img/oidc_auth_setting.png)

The settings of this auth mode:
* OIDC Provider Name: The name of the OIDC Provider.
* OIDC Provider Endpoint: The URL of the endpoint of the OIDC provider(a.k.a the Authorization Server in OAuth's terminology), 
which must service the "well-known" URI for its configuration, more details please refer to https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationRequest
* OIDC Client ID: The ID of client configured on OIDC Provider.
* OIDC Client Secret: The secret for this client.
* OIDC Scope: The scope values to be used during the authentication.  It is the comma separated string, which must contain `openid`.  
Normally it should also contain `profile` and `email`.  For getting the refresh token it should also contain `offline_access`.  Please check with the administrator of the OIDC Provider.
* Verify Certificate: Whether to check the certificate when accessing the OIDC Provider. if you are running the OIDC Provider with self-signed
certificate, make sure this value is set to false.


### Managing project creation
Use the **Project Creation** drop-down menu to set which users can create projects. Select **Everyone** to allow all users to create projects. Select **Admin Only** to allow only users with the Administrator role to create projects.  
![browse project](img/new_proj_create.png)

### Managing self-registration
You can manage whether a user can sign up for a new account. This option is not available if you use LDAP authentication.  
![browse project](img/new_self_reg.png)

### Managing email settings
You can change Harbor's email settings, the mail server is used to send out responses to users who request to reset their password.  
![browse project](img/new_config_email.png)

### Managing registry read only
You can change Harbor's registry read only settings, read only mode will allow 'docker pull' while preventing 'docker push' and the deletion of repository and tag.
![browse project](img/read_only.png)

If it set to true, deleting repository, tag and pushing image will be disabled. 
![browse project](img/read_only_enable.png)


```
$ docker push 10.117.169.182/demo/ubuntu:14.04  
The push refers to a repository [10.117.169.182/demo/ubuntu]
0271b8eebde3: Preparing 
denied: The system is in read only mode. Any modification is prohibited.  
```

### Managing role by LDAP group

If auth_mode is ldap_auth, you can manage project role by LDAP/AD group. please refer [manage role by ldap group guide](manage_role_by_ldap_group.md).

## Pulling and pushing images using Docker client  

**NOTE: Harbor only supports Registry V2 API. You need to use Docker client 1.6.0 or higher.**  

Harbor supports HTTP by default and Docker client tries to connect to Harbor using HTTPS first, so if you encounter an error as below when you pull or push images, you need to configure insecure registry. Please, read [this document](https://docs.docker.com/registry/insecure/) in order to understand how to do this. 

```
Error response from daemon: Get https://myregistrydomain.com/v1/users/: dial tcp myregistrydomain.com:443 getsockopt: connection refused.
```   

If this private registry supports only HTTP or HTTPS with an unknown CA certificate, please add   
`--insecure-registry myregistrydomain.com` to the daemon's start up arguments.  

In the case of HTTPS, if you have access to the registry's CA certificate, simply place the CA certificate at /etc/docker/certs.d/myregistrydomain.com/ca.crt .   

### Pulling images  
If the project that the image belongs to is private, you should sign in first:  

```
sh
$ docker login 10.117.169.182  
```

You can now pull the image:  

```
sh 
$ docker pull 10.117.169.182/library/ubuntu:14.04  
```

**Note**: Replace "10.117.169.182" with the IP address or domain name of your Harbor node. You cannot pull an unsigned image if you enabled content trust.

### Pushing images  
Before pushing an image, you must create a corresponding project on Harbor web UI. 

First, log in from Docker client:  

```
sh
$ docker login 10.117.169.182  
```

Tag the image:  

```
sh
$ docker tag ubuntu:14.04 10.117.169.182/demo/ubuntu:14.04  
```

Push the image:

```
sh
$ docker push 10.117.169.182/demo/ubuntu:14.04  
```

**Note: Replace "10.117.169.182" with the IP address or domain name of your Harbor node.**

###  Add description to repositories

After pushing an image, an Information can be added by project admin to describe this repository.

Go into the repository and select the "Info" tab, and click the "EDIT" button.  An textarea will appear and enter description here. Click "SAVE" button to save this information.

![edit info](img/edit_description.png)

### Download the harbor certs

Users  can click the "registry certificate" link to download the registry certificate.

![browse project](img/download_harbor_certs.png)

###  Deleting repositories  

Repository deletion runs in two steps.  

First, delete a repository in Harbor's UI. This is soft deletion. You can delete the entire repository or just a tag of it. After the soft deletion, 
the repository is no longer managed in Harbor, however, the files of the repository still remain in Harbor's storage.  

![browse project](img/new_delete_repo.png)
![browse project](img/new_delete_tag.png)

**CAUTION: If both tag A and tag B refer to the same image, after deleting tag A, B will also get deleted. if you enabled content trust, you need to use notary command line tool to delete the tag's signature before you delete an image.**  

Next, delete the actual files of the repository using the [garbage collection](#online-garbage-collection) in Harbor's UI. 

## Content trust  
**NOTE: Notary is an optional component, please make sure you have already installed it in your Harbor instance before you go through this section.**  
If you want to enable content trust to ensure that images are signed, please set two environment variables in the command line before pushing or pulling any image:
```sh
export DOCKER_CONTENT_TRUST=1
export DOCKER_CONTENT_TRUST_SERVER=https://10.117.169.182:4443
```
If you push the image for the first time, You will be asked to enter the root key passphrase. This will be needed every time you push a new image while the ``DOCKER_CONTENT_TRUST`` flag is set.  
The root key is generated at: ``/root/.docker/trust/private/root_keys``  
You will also be asked to enter a new passphrase for the image. This is generated at ``/root/.docker/trust/private/tuf_keys/[registry name] /[imagepath]``.  
If you are using a self-signed cert, make sure to copy the CA cert into ```/etc/docker/certs.d/10.117.169.182``` and ```$HOME/.docker/tls/10.117.169.182:4443/```. When an image is signed, it is indicated in the Web UI.  
**Note: Replace "10.117.169.182" with the IP address or domain name of your Harbor node. In order to use content trust, HTTPS must be enabled in Harbor.**  


When an image is signed, it has a tick shown in UI; otherwise, a cross sign(X) is displayed instead.  
![browse project](img/content_trust.png)

## Vulnerability Scanning

Harbor provides static analysis of vulnerabilities in images through the open source [Clair](https://github.com/coreos/clair) project. You can also connect Harbor to additional vulnerability scanners by using an interrogation service. The additional scanners that are currently supported are [Anchore](https://anchore.com/harbor/) and [Aqua](https://github.com/aquasecurity/harbor-scanner-aqua).

It might be necessary to connect Harbor to other scanners for corporate compliance reasons, or because your organization already uses a particular scanner. Different scanners also use different vulnerability databases, capture different CVE sets, and apply different severity thresholds. By connecting Harbor to more than one vulnerability scanner, you broaden the scope of your protection against vulnerabilities.

**IMPORTANT**: Clair is an optional component. To be able to use Clair or to configure additional scanners, you must have enabled Clair when you installed your Harbor instance. You configure additional scanners in the Harbor interface, after you have installed Harbor. 

- For information about installing Harbor with Clair, see the [Installation and Configuration Guide](installation_guide.md). 
- For information about adding additional scanners, see [Connect Harbor to Additional Vulnerability Scanners](#pluggable-scanners) below.

You can manually initiate scanning on a particular image, or on all images in Harbor. Additionally, you can also set a policy to automatically scan all of the images at specific intervals.

<a id="pluggable-scanners"></a>
### Connect Harbor to Additional Vulnerability Scanners

To connect Harbor to additional vulnerability scanners, you must have enabled the default Clair scanner when you deployed Harbor. You must install and configure an instance of the additional scanner according to the scanner vendor's requirements. The scanner must expose an API endpoint to Harbor that permits image pushes and pulls. You can deploy multiple different scanners, and multiple instances of the same type of scanner.

1. Log in to the Harbor interface with an account that has Harbor administrator privileges.
1. Expand **Administration**, and select **Interrogation Services**. 
   ![Interrogation Services](img/interrogation-services.png)
1. Click the **New Scanner** button.
1. Enter the information to identify the scanner.
   - A name for this scanner instance, to display in the Harbor interface.
   - An optional description of this scanner instance.
   - The address of the API endpoint that the scanner exposes to Harbor.
   ![Add scanner](img/add-scanner.png)
1. Select how to connect to the scanner from the **Authorization** drop-down menu.
   ![Select scanner authentication method](img/scanner-auth.png)
   - **None**: The scanner allows all connections without any security.
   - **Basic**: Enter a username and password for an account that can connect to the scanner.
   - **Bearer**: Paste the contents of a bearer token in the **Token** text box.
   - **APIKey**: Paste the contents of an API key for the scanner in the **APIKey** text box.
1. Optionally select **Skip certificate verification** if the scanner uses a self-signed or untrusted certificate. 
1. Optionally select **Use internal registry address** if the scanner should connect to Harbor using an internal network address rather its external URL.
1. Click **Test Connection** to make sure that Harbor can connect successfully to the scanner. 
   ![Test scanner connection](img/test-scanner-connection.png)
1. Click **Add** to connect Harbor to the scanner.
1. Optionally repeat the procedure to add more scanners.
1. If you configure multiple scanners, select one and click **Set as Default** to designate it as the default scanner.

### Updating the Vulnerability Metadata in the Database

Vulnerability scanners depend on the vulnerability metadata to complete the analysis process. After the first initial installation, the vulnerability scanner automatically starts to update the metadata database from different vulnerability repositories. The database update might take a while, based on the data size and network connection. 

Once the database is ready, the timestamp of the last update is shown in the **Interrogation Services** > **Vulnerability** tab. 
![browse project](img/clair_ready.png)

Until the database has been fully populated, the timestamp is replaced by a warning symbol. When the database is ready, you can scan images individually or scan all images across all projects.

### Scan Individual Images

1. Log in to the Harbor interface with an account that has at least project administrator privileges.
1. Go to **Projects** and select a project. 
1. Select the **Scanner** tab.

   The **Scanner** tab shows the details of the scanner that is currently set as the scanner to use for this project.
   
   ![Project scanner tab](img/project-scanners.png)
1. Click **Edit** to select a different scanner from the list of scanners that are connected to this Harbor instance, and click **OK**.

   ![Project scanner tab](img/select-scanner.png)
1. Select the **Repositories** tab and select a repository.

   For each tag in the repository, the **Vulnerabilities** column displays the vulnerability scanning status and related information.
   
   ![Tag vulnerability status](img/tag-vulnerability-status.png)
1. Select a tag, or use the check box at the top to select all tags in the repository, and click the **Scan** button to run the vulnerability scan on this image.

   ![Scan an image](img/scan_image.png)

   **NOTE**: You can start a scan at any time, unless the status is **Queued** or **Scanning**. If the database has not been fully populated, there is a warning message at the footer of the **Projects** > **Repositories** view and you cannot run the scan. The following statuses are displayed in the **Vulnerabilities** column:
   
   * **Not Scanned:** The tag has never been scanned.
   * **Queued:** The scanning task is scheduled but has not run yet.
   * **Scanning:** The scanning task is in progress and a progress bar is displayed.
   * **View log:** The scanning task failed to complete. Click **View Log** link to view the related logs.
   * **Complete:** The scanning task completed successfully.

   If the process completes successfully, the result indicates the number of vulnerabilities with each severity level and the number of vulnerabilities that are fixable.

   ![Scan result](img/scan-result.png)

   * **Red:** Critical level of vulnerabilities
   * **Orange:** High level of vulnerabilities
   * **Yellow:** Medium level of vulnerabilities
   * **Blue:** Low level of vulnerabilities
   * **Green:** No vulnerability
   * **Grey:** Unknown level of vulnerabilities
1. Hover over the number of fixable vulnerabilities to see a summary of the vulnerability report. 

   ![Vulnerability summary](img/vulnerability-summary.png)
1. Click on the tag name to see a detailed vulnerability report.
 
  ![Vulnerability report](img/tag_detail.png)
  
   In addition to information about the tag, all of the vulnerabilities found in the last scan are listed. You can order or filter the list by the different columns.

### Scan All Images

In addition to scanning individual images in projects, you can run global scans on all of the images in a Harbor instance, across all projects.

1. Log in to the Harbor interface with an account that has Harbor administrator privileges.
1. Expand **Administration**, and select **Interrogation Services**. 
1. Select the **Vulnerability** tab and click **Scan Now** to scan all of the images in all projects.

   ![Scan all images](img/scan_all.png)
   
Scanning requires intensive resource consumption. To avoid frequently triggering scans too frequently, scans can be only triggered once in a defined period. If scanning is unavailable, the next available time is displayed next to the **Scan Now** button.

### Schedule Scans

You can set policies to control when vulnerability scanning should run.

1. Log in to the Harbor interface with an account that has Harbor administrator privileges.
1. Expand **Administration**, and select **Interrogation Services**. 
1. Select the **Vulnerability** tab and click the **Edit** button next to **Schedule to scan all**.  
1. Use the drop down-menu to select how often to run scans.

   ![browse project](img/scan_policy.png)
   
   * **None**: No scans are scheduled.
   * **Hourly**: Run a scan at the beginning of every hour.
   * **Daily**: Run a scan at midnight every day.
   * **Weekly**: Run a scan at midnight every Saturday.
   * **Custom**: Run a scan according to a `cron` job.
1. Click **Save**.

## Configure CVE Whitelists

When you run vulnerability scans, images that are subject to Common Vulnerabilities and Exposures (CVE) are identified. According to the severity of the CVE and your security settings, these images might not be permitted to run. As a system administrator, you can create whitelists of CVEs to ignore during vulnerability scanning. 

You can set a system-wide CVE whitelist or you can set CVE whitelists on a per-project basis.

### Configure a System-Wide CVE Whitelist

System-wide CVE whitelists apply to all of the projects in a Harbor instance.

1. Go to **Configuration** > **System Settings**.
1. Under **Deployment security**, click **Add**. 
   ![System-wide CVE whitelist](img/cve-whitelist1.png)
1. Enter the list of CVE IDs to ignore during vulnerability scanning. 
   ![Add system CVE whitelist](img/cve-whitelist2.png)

   Either use a comma-separated list or newlines to add multiple CVE IDs to the list.
1. Click **Add** at the bottom of the window to add the list.
1. Optionally uncheck the **Never expires** checkbox and use the calendar selector to set an expiry date for the whitelist.
   ![Add system CVEs](img/cve-whitelist3.png)
1. Click **Save** at the bottom of the page to save your settings.
   
After you have created a system whitelist, you can remove CVE IDs from the list by clicking the delete button next to it in the list. You can click **Add** to add more CVE IDs to the system whitelist.

![Add and remove system CVEs](img/cve-whitelist4.png)

### Configure a Per-Project CVE Whitelist

By default, the system whitelist is applied to all projects. You can configure different CVE whitelists for individual projects, that override the system whitelist. 

1. Go to **Projects**, select a project, and select **Configuration**.
1. Under **CVE whitelist**, select **Project whitelist**. 
   ![Project CVE whitelist](img/cve-whitelist5.png)
1. Optionally click **Copy From System** to add all of the CVE IDs from the system CVE whitelist to this project whitelist.
1. Click **Add** and enter a list of additional CVE IDs to ignore during vulnerability scanning of this project. 
   ![Add project CVEs](img/cve-whitelist6.png)

   Either use a comma-separated list or newlines to add multiple CVE IDs to the list.
1. Click **Add** at the bottom of the window to add the CVEs to the project whitelist.
1. Optionally uncheck the **Never expires** checkbox and use the calendar selector to set an expiry date for the whitelist.
1. Click **Save** at the bottom of the page to save your settings.

After you have created a project whitelist, you can remove CVE IDs from the list by clicking the delete button next to it in the list. You can click **Add** at any time to add more CVE IDs to this project whitelist. 

If CVEs are added to the system whitelist after you have created a project whitelist, click **Copy From System** to add the new entries from the system whitelist to the project whitelist. 

**NOTE**: If CVEs are deleted from the system whitelist after you have created a project whitelist, and if you added the system whitelist to the project whitelist, you must manually remove the deleted CVEs from the project whitelist. If you click **Copy From System** after CVEs have been deleted from the system whitelist, the deleted CVEs are not automatically removed from the project whitelist.

## Pull image from Harbor in Kubernetes
Kubernetes users can easily deploy pods with images stored in Harbor.  The settings are similar to that of another private registry. There are two major issues:

1. When your Harbor instance is hosting http and the certificate is self signed.  You need to modify daemon.json on each work node of your cluster, for details please refer to: https://docs.docker.com/registry/insecure/#deploy-a-plain-http-registry
2. If your pod references an image under private project, you need to create a secret with the credentials of user who has permission to pull image from this project, for details refer to: https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/

## Manage Helm Charts
[Helm](https://helm.sh) is a package manager for [Kubernetes](https://kubernetes.io). Helm uses a packaging format called [charts](https://docs.helm.sh/developing_charts). Since version 1.6.0 Harbor is now a composite cloud-native registry which supports both container image management and Helm charts management. Access to Helm charts in Harbor is controlled by [role-based access controls (RBAC)](https://en.wikipedia.org/wiki/Role-based_access_control) and is restricted by projects.

### Manage Helm Charts via portal
#### List charts
Click your project to enter the project detail page after successful logging in. The existing helm charts will be listed under the tab `Helm Charts` which is beside the image `Repositories` tab with the following information:
* Name of helm chart
* The status of the chart: Active or Deprecated
* The count of chart versions
* The created time of the chart

![list charts](img/chartrepo/list_charts.png)

You can click the icon buttons on the top right to switch views between card view and list view.

#### Upload new chart
Click the `UPLOAD` button on the top left to open the chart uploading dialog. Choose the uploading chart from your filesystem. Click the `UPLOAD` button to upload it to the chart repository server.

![upload charts](img/chartrepo/upload_charts.png)

If the chart is signed, you can choose the corresponding provenance file from your filesystem and Click the `UPLOAD` button to upload them together at once.

If the chart is successfully uploaded, it will be displayed in the chart list at once.

#### List chart versions
Clicking the chart name from the chart list will show all the available versions of that chart with the following information:
* the chart version number
* the maintainers of the chart version
* the template engine used (default is gotpl)
* the created timestamp of the chart version

![list charts versions](img/chartrepo/list_chart_versions.png)

Obviously, there will be at least 1 version for each of the charts in the top chart list. Same with chart list view, you can also click the icon buttons on the top right to switch views between card view and list view.

Check the checkbox at the 1st column to select the specified chart versions:
* Click the `DELETE` button to delete all the selected chart versions from the chart repository server. Batch operation is supported.
* Click the `DOWNLOAD` button to download the chart artifact file. Batch operation is not supported.
* Click the `UPLOAD` button to upload the new chart version for the current chart

#### Adding labels to/remove labels from chart versions
Users who have system administrator, project administrator or project developer role can click the `ADD LABELS` button to add labels to or remove labels from chart versions.

![add labels to chart versions](img/chartrepo/add_labesl_to_chart_versions.png)


#### Filtering chart versions by labels
The chart versions can be filtered by labels:

![filter chart versions by labels](img/chartrepo/filter_chart_versions_by_label.png)

#### View chart version details
Clicking the chart version number link will open the chart version details view. You can see more details about the specified chart version here. There are three content sections:
* **Summary:**
  * readme of the chart
  * overall metadata like home, created timestamp and application version
  * related helm commands for reference, such as `helm add repo` and `helm install` etc.
![chart details](img/chartrepo/chart_details.png)
* **Dependencies:**
  * list all the dependant sun charts with 'name', 'version' and 'repository' fields
![chart dependencies](img/chartrepo/chart_dependencies.png)
* **Values:**
  * display the content from `values.yaml` file with highlight code preview
  * clicking the icon buttons on the top right to switch the yaml file view to k-v value pair list view
![chart values](img/chartrepo/chart_values.png)

Clicking the `DOWNLOAD` button on the top right will start the downloading process.

### Working with Helm CLI
As a helm chart repository, Harbor can work smoothly with Helm CLI. About how to install Helm CLI, please refer [install helm](https://docs.helm.sh/using_helm/#installing-helm). Run command `helm version` to make sure the version of Helm CLI is v2.9.1+.
```
helm version

#Client: &version.Version{SemVer:"v2.9.1", GitCommit:"20adb27c7c5868466912eebdf6664e7390ebe710", GitTreeState:"clean"}
#Server: &version.Version{SemVer:"v2.9.1", GitCommit:"20adb27c7c5868466912eebdf6664e7390ebe710", GitTreeState:"clean"}
```
#### Add harbor to the repository list
Before working, Harbor should be added into the repository list with `helm repo add` command. Two different modes are supported.
* Add Harbor as a unified single index entry point

With this mode Helm can be made aware of all the charts located in different projects and which are accessible by the currently authenticated user.
```
helm repo add --ca-file ca.crt --username=admin --password=Passw0rd myrepo https://xx.xx.xx.xx/chartrepo
```
**NOTES:** Providing both ca file and cert files is caused by an issue from helm.

* Add Harbor project as separate index entry point

With this mode, helm can only pull charts in the specified project.
```
helm repo add --ca-file ca.crt --username=admin --password=Passw0rd myrepo https://xx.xx.xx.xx/chartrepo/myproject
```

#### Push charts to the repository server by CLI
As an alternative, you can also upload charts via the CLI. It is not supported by the native helm CLI. A plugin from the community should be installed before pushing. Run `helm plugin install` to install the `push` plugin first.
```
helm plugin install https://github.com/chartmuseum/helm-push
```
After a successful installation,  run `push` command to upload your charts:
```
helm push --ca-file=ca.crt --username=admin --password=passw0rd chart_repo/hello-helm-0.1.0.tgz myrepo
```
**NOTES:** `push` command does not support pushing a prov file of a signed chart yet.

#### Install charts
Before installing, make sure your helm is correctly initialized with command `helm init` and the chart index is synchronized with command `helm repo update`.

Search the chart with the keyword if you're not sure where it is:
```
helm search hello

#NAME                            CHART VERSION   APP VERSION     DESCRIPTION                
#local/hello-helm                0.3.10          1.3             A Helm chart for Kubernetes
#myrepo/chart_repo/hello-helm    0.1.10          1.2             A Helm chart for Kubernetes
#myrepo/library/hello-helm       0.3.10          1.3             A Helm chart for Kubernetes
```
Everything is ready, install the chart to your kubernetes:
```
helm install --ca-file=ca.crt --username=admin --password=Passw0rd --version 0.1.10 repo248/chart_repo/hello-helm
```

For other more helm commands like how to sign a chart, please refer to the [helm doc](https://docs.helm.sh/helm/#helm).

## Garbage Collection

When you delete images from Harbor, space is not automatically freed up. You must run garbage collection to free up space by removing blobs that are no longer referenced by a manifest from the file system.  

### Run Garbage Collection

1. Log in to the Harbor interface with an account that has Harbor administrator privileges.
1. Expand **Administration**, and select **Garbage Collection**. 
1. Select the **'Garbage Collection'** tab.
  ![Garbage collection](img/garbage-collection.png)
1. To run garbage collection immediately, click **GC Now**.

When you run garbage collection, Harbor goes into read-only mode. All modifications to the registry are prohibited.

To avoid triggering the garbage collection process too frequently, the availability of the **GC Now** button is restricted. Garbage collection can be only run once per minute.

### Schedule Garbage Collection

1. Expand **Administration**, and select **Garbage Collection**. 
1. Select the **'Garbage Collection'** tab.
1. Use the drop down-menu to select how often to run garbage collection.
   ![Schedule garbage collection](img/gc_policy.png)
   * **None**: No garbage collection is scheduled.
   * **Hourly**: Run garbage collection at the beginning of every hour.
   * **Daily**: Run garbage collection at midnight every day.
   * **Weekly**: Run garbage collection at midnight every Saturday.
   * **Custom**: Run garbage collection according to a `cron` job.
1. Click **Save**.
1. Select the **History** tab to view records of the 10 most recent garbage collection runs.
   ![Garbage collection history](img/gc_history.png)
1. Click on the **Logs** link to view the related logs.

## Build history

Build history make it easy to see the contents of a container image, find the code which bulids an image, or locate the image for a source repository.

In Harbor portal, enter your project, select the repository, click on the link of tag name you'd like to see its build history, the detail page will be opened. Then switch to `Build History` tab, you can see the build history information.

![build_ history](img/build_history.png)

## Using OIDC CLI secret

Having authenticated via OIDC SSO and onboarded to Harbor, you can use Docker/Helm CLI to access Harbor to read/write the artifacts.
As the CLI cannot handle redirection for SSO, we introduced `CLI secret`, which is only available when Harbor's authentication mode 
is configured to OIDC based.  
After logging into Harbor, click the drop down list to view user's profile:
![user_profile](img/user_profile.png)

You can copy your CLI secret via the dialog of profile:
![profile_dlg](img/profile_dlg.png)

After that you can authenticate using your user name in Harbor that you set during onboard process, and CLI secret as the password
with Docker/Helm CLI, for example:
```sh
docker login -u testuser -p xxxxxx jt-test.local.goharbor.io

``` 

When you click the "..." icon in the profile dialog, a button for generating new CLI secret will appear, and you can generate a new 
CLI secret by clicking this button.  Please be reminded one user can only have one CLI secret, so when a new secret is generated, the
old one becomes invalid at once.

**NOTE**:
Under the hood the CLI secret is associated with the ID token, and Harbor will try to refresh the token, so the CLI secret will
be valid after th ID token expires. However, if the OIDC Provider does not provide refresh token or the refresh fails for some 
reason, the CLI secret will become invalid.  In that case you can logout and login Harbor via SSO flow again so Harbor can get a 
new ID token and the CLI secret will work again.


## Robot Account
Robot Accounts are accounts created by project admins that are intended for automated operations. They have the following limitations:

1, Robot Accounts cannot login Harbor portal
2, Robot Accounts can only perform operations by using the Docker and Helm CLIs.

### Add a Robot Account
If you are a project admin, you can create a Robot Account by clicking "New Robot Account" in the `Robot Accounts` tab of a project, and enter a name, a description, and grant permission to the account to push and pull images and Helm charts.
![add_robot_account](img/robotaccount/add_robot_account.png)

![add_robot_account](img/robotaccount/add_robot_account_2.png)

> **NOTE:** The name will become `robot$<accountname>` and will be used to distinguish a robot account from a normal harbor user.

![copy_robot_account_token](img/robotaccount/copy_robot_account_token.png)
As Harbor doesn't store your account token, please make sure to copy it in the pop up dialog after creating, otherwise, there is no way to get it from Harbor.

### Configure duration of robot account
If you are a system admin, you can configure the robot account token duration in days. 
![set_robot_account_token_duration](img/robotaccount/set_robot_account_token_duration.png)

### Authenticate with a robot account
To authenticate with a Robot Account, use `docker login` as below,

```
docker login harbor.io
Username: robot$accountname
Password: Thepasswordgeneratedbyprojectadmin
```

### Disable a robot account
If you are a project admin, you can disable a Robot Account by clicking "Disable Account" in the `Robot Accounts` tab of a project.
![disable_robot_account](img/robotaccount/disable_delete_robot_account.png)

### Delete a robot account
If you are a project admin, you can delete a Robot Account by clicking "Delete" in the `Robot Accounts` tab of a project.
![delete_robot_account](img/robotaccount/disable_delete_robot_account.png)

## Tag Retention Rules

A repository can rapidly accumulate a large number of image tags, many of which might not be required after a given time or once they have been superseded by a subsequent image build. These excess tags can obviously consume large quantities of storage capacity. As a system administrator, you can define rules that govern how many tags of a given repository to retain, or for how long to retain certain tags. 

### How Tag Retention Rules Work

You define tag retention rules on repositories, not on projects. This allows for greater granularity when defining your retention rules. As the name suggests, when you define a retention rule for a repository, you are identifying which tags to retain. You do not define rules to explicitly remove tags. Rather, when you set a rule, any tags in a repository that are not identified as being eligible for retention are discarded. 

A tag retention rule has 3 filters that are applied sequentially, as described in the following table.

|Order|Filter|Description|
|---|---|---|
|First|Repository or repositories|Identify the repository or repositories on which to apply the rule. You can identify repositories that either have a certain name or name fragment, or that do not have that name or name fragment. Wild cards (for example `*repo`, `repo*`, and `**`) are permitted. The repository filter is applied first to mark the repositories to which to apply the retention rule. The identified repositories are earmarked for further matching based on the tag criteria. No action is taken on the nonspecified repositories at this stage.|
|Second|Quantity to retain|Set which tags to retain either by specifying a maximum number of tags, or by specifying a maximum period for which to retain tags.|
|Third|Tags to retain|Identify the tag or tags on which to apply the rule. You can identify tags that either have a certain name or name fragment, or that do not have that name or name fragment. Wild cards (for example `*tag`, `tag*`, and `**`) are permitted.|

For information about how the `**` wildcard is applied, see https://github.com/bmatcuk/doublestar#patterns.

#### Example 1

- You have 5 repositories in a project, repositories A to E.
  - Repository A has 100 image tags, all of which have been pulled in the last week.
  - Repositories B to E each have 6 images, none of which have been pulled in the last month.
- You set the repository filter to `**`, meaning that all repositories in the project are included.
- You set the retention policy to retain the 10 most recently pulled images in each repository.
- You set the tag filter to `**`, meaning that all tags in the repository are included.

In this example the rule retains the 10 most recently pulled images in repository A, and all 6 of the images in each of the 4 repositories B to E. So, a total of 34 image tags are retained in the project. In other words, the rule does not treat all of the images in repositories A to E as a single pool from which to choose the 10 most recent images. So, even if the 11th to 100th tags in repository A have been pulled more recently than any of the tags in repositories B to E, all of the tags in repositories B to E are retained, because each of those repositories has fewer than 10 tags.

#### Example 2

This example uses the same project and repositories as example 1, but sets the retention policy to retain the images in each repository that have been pulled in the last 7 days.

In this case, all of the images in repository A are retained because they have been pulled in the last 7 days. None of the images in repositories B to E are retained, because none of them has been pulled in the last week. In this example, 100 images are retained, as opposed to 34 images in example 1.

#### Tag Retention Rules and Native Docker Tag Deletion

**WARNING**: Due to native Docker tag deletion behavior, there is an issue with the current retention policy implementation. If you have multiple tags that refer to the same SHA digest, and if a subset of these tags are marked for deletion by a configured retention policy, all of the remaining tags would also be deleted. This violates the retention policy, so in this case all of the tags are retained. This issue will be addressed in a future update release, so that tag retention policies can delete tags without deleting the digest and other shared tags.

For example, you have following tags, listed according to their push time, and all of them refer to the same SHA digest:

- `harbor-1.8`, pushed 8/14/2019 12:00am
- `harbor-release`, pushed 8/14/2019 03:00am
- `harbor-nightly`, pushed 8/14/2019 06:00am
- `harbor-latest`, pushed 8/14/2019 09:00am

You configure a retention policy to retain the two latest tags that match `harbor-*`, so that `harbor-rc` and `harbor-latest` are deleted. However, since all tags refer to the same SHA digest, this policy would also delete the tags `harbor-1.8` and `harbor-release`, so all tags are retained.

### Combining Rules on a Repository

You can define up to 15 rules per project. You can apply multiple rules to a repository or set of repositories. When you apply multiple rules to a repository, they are applied with `OR` logic rather than with `AND` logic. In this way, there is no prioritization of application of the rules on a given repository. Rules run concurrently in the background, and the resulting sets from each rule are combined at the end of the run.

#### Example 3

This example uses the same project and repositories as examples 1 and 2, but sets two rules:

- Rule 1: Retain all of the images in each repository that have been pulled in the last 7 days.
- Rule 2: Retain a maximum number of 10 images in each repository.

For repository A, rule 1 retains all of the images because they have all been pulled in the last week. Rule 2 retains the 10 most recently pulled images. So, since the two rules are applied with an `OR` relationship, all 100 images are retained in repository A.

For repositories B-E, rule 1 will retain 0 images as no images are pulled in the last week. Rule 2 will retain all 6 images because 6 < 10. So, since the two rules are applied with an `OR` relationship, for repositories B-E, each repository will keep all 6 images.

In this example, all of the images are retained.

#### Example 4

This example uses a different repository to the previous examples.

- You have a repository that has 12 tags:

  |Production|Release Candidate|Release|
  |---|---|---|
  |`2.1-your_repo-prod`|`2.1-your_repo-rc`|`2.1-your_repo-release`|
  |`2.2-your_repo-prod`|`2.2-your_repo-rc`|`2.2-your_repo-release`|
  |`3.1-your_repo-prod`|`3.1-your_repo-rc`|`3.1-your_repo-release`|
  |`4.4-your_repo-prod`|`4.4-your_repo-rc`|`4.4-your_repo-release`| 

- You define three tag retention rules on this repository:
  - Retain the 10 most recently pushed image tags that start with `2`.
  - Retain the 10 most recently pushed image tags that end with `-prod`.
  - Retain all tags that do not include `2.1-your_repo-prod`.

In this example, the rules are applied to the following 7 tags:

- `2.1-your_repo-rc`
- `2.1-your_repo-release`
- `2.2-your_repo-prod`
- `2.2-your_repo-rc`
- `2.2-your_repo-release`
- `3.1-your_repo-prod`
- `4.4-your_repo-prod`

### How Tag Retention Rules Interact with Project Quotas

The system administrator can set a maximum on the number of tags that a project can contain and the amount of storage that it can consume. For information about project quotas, see [Set Project Quotas](#set-project-quotas). 

If you set a quota on a project, this quota cannot be exceeded. The quota is applied to a project even if you set a retention rule that would exceed it. In other words, you cannot use retention rules to bypass quotas.

### Configure Tag Retention Rules

1. Log in to the Harbor interface with an account that has at least project administrator privileges.
1. Go to **Projects**, select a project, click the **More** ellipsis (`...`), and select **Tag Strategy**.

   ![Tag Retention option](img/tag-retention1.png)
1. Select **Tag Retention**.

   ![Tag Retention option](img/tag-retention1a.png)
  
1. Click **Add Rule** to add a rule.
1. In the **For the repositories** drop-down menu, select **matching** or **excluding**.
  ![Select repositories](img/tag-retention2.png)
1. In the **For the repositories** text box, identify the repositories on which to apply the rule.
  
   You can define the repositories on which to apply the rule by entering the following information:
  
   - A repository name, for example `my_repo_1`.
   - A comma-separated list of repository names, for example `my_repo_1,my_repo_2,your_repo_3`.
   - A partial repository name with wildcards, for example `my_*`, `*_3`, or `*_repo_*`.
   - `**` to apply the rule to all of the repositories in the project. 
  
   If you selected **matching**, the rule is applied to the repositories you identified. If you selected **excluding**, the rule is applied to all of the repositories in the project except for the ones that you identified.
1. In the **By image count or number of days** drop-down menu, define how many tags to retain or the period to retain tags.
  ![Select retention criteria](img/tag-retention3.png)
  
   |Option|Description|
   |---|---|
   |**retain the most recently pushed # images**|Enter the maximum number of images to retain, keeping the ones that have been pushed most recently. There is no maximum age for an image.|
   |**retain the most recently pulled # images**|Enter the maximum number of images to retain, keeping only the ones that have been pulled recently. There is no maximum age for an image.|
   |**retain the images pushed within the last # days**|Enter the number of days to retain images, keeping only the ones that have been pushed during this period. There is no maximum number of images.|
   |**retain the images pulled within the last # days**|Enter the number of days to retain images, keeping only the ones that have been pulled during this period. There is no maximum number of images.|
   |**retain always**|Always retain the images identified by this rule.| 

1. In the **Tags** drop-down menu, select **matching** or **excluding**.
1. In the **Tags** text box, identify the tags on which to apply the rule.
  
   You can define the tags on which to apply the rule by entering the following information:
  
   - A tag name, for example `my_tag_1`.
   - A comma-separated list of tag names, for example `my_tag_1,my_tag_2,your_tag_3`.
   - A partial tag name with wildcards, for example `my_*`, `*_3`, or `*_tag_*`.
   - `**` to apply the rule to all of the tags in the project. 
  
   If you selected **matching**, the rule is applied to the tags you identified. If you selected **excluding**, the rule is applied to all of the tags in the repository except for the ones that you identified.
1. Click **Add** to save the rule.
1. (Optional) Click **Add Rule** to add more rules, up to a maximum of 15 per project.
1. (Optional) Under Schedule, click **Edit** and select how often to run the rule.

   ![Select retention criteria](img/tag-retention4.png)
   
   If you select **Custom**, enter a cron job command to schedule the rule. 
  
   **NOTE**: If you define multiple rules, the schedule is applied to all of the rules. You cannot schedule different rules to run at different times. 
1. Click **Dry Run** to test the rule or rules that you have defined.
1. Click **Run Now** to run the rule immediately.

**WARNING**: You cannot revert a rule after you run it. It is strongly recommended to perform a dry run before you run rules. 

To modify an existing rule, use the **Action** drop-down menu next to a rule to disable, edit, or delete that rule. 

![Modify tag retention rules](img/tag-retention5.png)

## Configure Tag Immutability

By default, users can repeatedly push an image with the same tag to repositories in Harbor. This causes the previous image to effectively be overwritten with each push, in that the tag now points to a different image and the image that previously used the tag now becomes tagless. This is due to the Docker implementation, that does not enforce the mapping between an image tag and the image digest. This can be undesirable in certain cases, because the tag can no longer be trusted to identify the image version. The sha256 digest remains reliable and always points to the same build, but it is not rendered in a human-friendly format.  

Moreover, the Docker implementation requires that deleting a tag results in the deletion of all other tags that point to the same digest, causing unwanted image deletions.

To prevent this, Harbor allows you to configure tag immutability at the project level, so that images with certain tags cannot be pushed into Harbor if their tags match existing tags. This prevents existing images from being overwritten. Tag immutability guarantees that an immutable tagged image cannot be deleted, and cannot be altered through repushing, retagging, or replication.

1. Log in to the Harbor interface with an account that has at least project administrator privileges.
1. Go to **Projects**, select a project, click the **More** ellipsis (`...`), and select **Tag Strategy**.

   ![Tag Retention option](img/tag-retention1.png)
1. Select **Tag Immutability**.

   ![Add an immutability rule](img/tag-immutability.png)
1. Click **Add Rule**.

   - In the **For the respositories** row, enter a comma-separated list of repositories to which to either apply or exclude from the rule by selecting either **matching** or **excluding** from the drop-down menu.
   - In the **Tags** row, enter a comma-separated list of tags to which to either apply or exclude from the rule by selecting either **matching** or **excluding** from the drop-down menu.
 
   ![Add an immutability rule](img/add-immutability-rule.png)
1. Click **Add** to save the rule.

For example, to make all tags for all repositories in the project immutable, set the following options:

- Set **For the respositories** to **matching** and enter `**`.
- Set **Tags** to **matching** and enter `**`.

To allow the tags `rc`, `test`, and `nightly` to be overwritten but make all other tags immutable, set the following options:

- Set **For the respositories** to **matching** and enter `**`.
- Set **Tags** to **excluding** and enter `rc,test,nightly`.

To modify an existing rule, use the **Action** drop-down menu next to a rule to disable, edit, or delete that rule. 

![Modify tag retention rules](img/edit-tag-immutability.png)

## Webhook Notifications

If you are a project administrator, you can configure a connection from a project in Harbor to a webhook endpoint. If you configure webhooks, Harbor notifies the webhook endpoint of certain events that occur in the project. Webhooks allow you to integrate Harbor with other tools to streamline continuous integration and development processes. 

The action that is taken upon receiving a notification from a Harbor project depends on your continuous integration and development processes. For example, by configuring Harbor to send a `POST` request to a webhook listener at an endpoint of your choice, you can trigger a build and deployment of an application whenever there is a change to an image in the repository.

### Supported Events

You can define one webhook endpoint per project. Webhook notifications provide information about events in JSON format and are delivered by `HTTP` or `HTTPS POST` to an existing webhhook endpoint URL that you provide. The following table describes the events that trigger notifications and the contents of each notification.

|Event|Webhook Event Type|Contents of Notification|
|---|---|---|
|Push image to registry|`IMAGE PUSH`|Repository namespace name, repository name, resource URL, tags, manifest digest, image name, push time timestamp, username of user who pushed image|
|Pull manifest from registry|`IMAGE PULL`|Repository namespace name, repository name, manifest digest, image name, pull time timestamp, username of user who pulled image|
|Delete manifest from registry|`IMAGE DELETE`|Repository namespace name, repository name, manifest digest, image name, image size, delete time timestamp, username of user who deleted image|
|Upload Helm chart to registry|`CHART PUSH`|Repository name, chart name, chart type, chart version, chart size, tag, timestamp of push, username of user who uploaded chart|
|Download Helm chart from registry|`CHART PULL`|Repository name, chart name, chart type, chart version, chart size, tag, timestamp of push, username of user who pulled chart|
|Delete Helm chart from registry|`CHART DELETE`|Repository name, chart name, chart type, chart version, chart size, tag, timestamp of delete, username of user who deleted chart|
|Image scan completed|`IMAGE SCAN COMPLETED`|Repository namespace name, repository name, tag scanned, image name, number of critical issues, number of major issues, number of minor issues, last scan status, scan completion time timestamp, vulnerability information (CVE ID, description, link to CVE, criticality, URL for any fix), username of user who performed scan|
|Image scan failed|`IMAGE SCAN FAILED`|Repository namespace name, repository name, tag scanned, image name, error that occurred, username of user who performed scan|
|Project quota exceeded|`PROJECT QUOTA EXCEED`|Repository namespace name, repository name, tags, manifest digest, image name, push time timestamp, username of user who pushed image|

#### JSON Payload Format

The webhook notification is delivered in JSON format. The following example shows the JSON notification for a push image event:

```
{
 "event_type": "pushImage"
    "events": [
               {
                "project": "prj",
                "repo_name": "repo1",
                "tag": "latest",
                "full_name": "prj/repo1",
                "trigger_time": 158322233213,
                "image_id": "9e2c9d5f44efbb6ee83aecd17a120c513047d289d142ec5738c9f02f9b24ad07",
                "project_type": "Private"
               }
             ]
}
```

### Webhook Endpoint Recommendations

The endpoint that receives the webhook should ideally have a webhook listener that is capable of interpreting the payload and acting upon the information it contains. For example, running a shell script.

### Example Use Cases

You can configure your continuous integration and development infrastructure so that it performs the following types of operations when it receives a webhook notification from Harbor.

- Image push: 
  - Trigger a new build immediately following a push on selected repositories or tags.
  - Notify services or applications that use the image that a new image is available and pull it.
  - Scan the image using Clair.
  - Replicate the image to remote registries.
- Image scanning:
  - If a vulnerability is found, rescan the image or replicate it to another registry.
  - If the scan passes, deploy the image.

### Configure Webhooks

1. Select a project and go to the Webhooks tab.
  ![Webhooks option](img/webhooks1.png)  
1. Enter the URL for your webhook endpoint listener.
  ![Webhook URL](img/webhooks2.png)
1. If your webhook listener implements authentication, enter the authentication header. 
1. To implement `HTTPS POST` instead of `HTTP POST`, select the **Verifiy Remote Certficate** check box.
1. Click **Test Endpoint** to make sure that Harbor can connect to the listener.
1. Click **Continue** to create the webhook.

When you have created the webhook, you see the status of the different notifications and the timestamp of the last time each notification was triggered. You can click **Disable** to disable notifications.

![Webhook Status](img/webhooks3.png)

**NOTE**: You can only disable and reenable all notifications. You cannot disable and enable selected notifications.

If a webhook notification fails to send, or if it receives an HTTP error response with a code other than `2xx`, the notification is re-sent based on the configuration that you set in `harbor.yml`. 

### Globally Enable and Disable Webhooks

As a system administrator, you can enable and disable webhook notifications for all projects.

1. Go to **Configuration** > **System Settings**.
1. Scroll down and check or uncheck the **Webhooks enabled** check box.

  ![Enable/disable webhooks](img/webhooks4.png)

## API Explorer

Harbor integrated swagger UI from 1.8. That means all apis can be invoked through UI. Normally, user have 2 ways to navigate to API Explorer. 

1. User can login harbor, and click the "API EXPLORER" button.All apis will be invoked with current user authorization.                         
![navigation bar](img/api_explorer_btn.png)  


2. User can navigate to swagger page by ip address by router "devcenter". For example: https://10.192.111.118/devcenter. After go to the page, need to click "authorize" button to give basic authentication to all apis. All apis will be invoked with the authorized user authorization. 
![authentication](img/authorize.png)


