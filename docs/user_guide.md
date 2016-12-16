#User Guide
##Overview
This guide walks you through the fundamentals of using Harbor. You'll learn how to use Harbor to:  

* Manage your projects.
* Manage members of a project.
* Replicate projects to a remote registry.
* Search projects and repositories.
* Manage Harbor system if you are the system administrator:
 + Manage users.
 + Manage destinations.
 + Manage replication policies.
* Pull and push images using Docker client.
* Delete repositories and images.


##Role Based Access Control

![rbac](img/rbac.png)

Harbor manages images through projects. Users can be added into one project as a member with three different roles:  

* **Guest**: Guest has read-only privilege for a specified project.
* **Developer**: Developer has read and write privileges for a project.
* **ProjectAdmin**: When creating a new project, you will be assigned the "ProjectAdmin" role to the project. Besides read-write privileges, the "ProjectAdmin" also has some management privileges, such as adding and removing members.

Besides the above three roles, there are two system-wide roles:  

* **SysAdmin**: "SysAdmin" has the most privileges. In addition to the privileges mentioned above, "SysAdmin" can also list all projects, set an ordinary user as administrator and delete users. The public project "library" is also owned by the administrator.  
* **Anonymous**: When a user is not logged in, the user is considered as an "anonymous" user. An anonymous user has no access to private projects and has read-only access to public projects.  

##User account
Harbor supports two authentication modes:  

* **Database(db_auth)**  

	Users are stored in the local database.  
	
	A user can self register himself/herself in Harbor in this mode. To disable user self-registration, refer to the [installation guide](installation_guide_ova.md). When self-registration is disabled, the system administrator can add users in Harbor.  
	
	When registering or adding a new user, the username and email must be unique in the Harbor system. The password must contain at least 8 characters with 1 lowercase letter, 1 uppercase letter and 1 numeric character.  
	
	When you forgot your password, you can follow the below steps to reset the password:  

	1. Click the link "Forgot Password" in the sign in page.  
	2. Input the email address entered when you signed up, an email will be sent out to you for password reset.  
	3. After receiving the email, click on the link in the email which directs you to a password reset web page.  
	4. Input your new password and click "Save".  
	
* **LDAP/Active Directory (ldap_auth)**  

	Under this authentication mode, users whose credentials are stored in an external LDAP or AD server can log in to Harbor directly.  
	
	When an LDAP/AD user logs in by *username* and *password*, Harbor binds to the LDAP/AD server with the **"LDAP Search DN"** and **"LDAP Search Password"** described in [installation guide](installation_guide_ova.md). If it successes, Harbor looks up the user under the LDAP entry **"LDAP Base DN"** including substree. The attribute (such as uid, cn) specified by **"LDAP UID"** is used to match a user with the *username*. If a match is found, the user's *password* is verified by a bind request to the LDAP/AD server.  
	
	Self-registration, changing password and resetting password are not supported anymore under LDAP/AD authentication mode because the users are managed by LDAP or AD.  

##Managing projects
A project in Harbor contains all repositories of an application. No images can be pushed to Harbor before the project is created. RBAC is applied to a project. There are two types of projects in Harbor:  

* **Public**: All users have the read privilege to a public project, it's convenient for you to share some repositories with others in this way.
* **Private**: A private project can only be accessed by users with proper privileges.  

You can create a project after you signed in. Enabling the "Public" checkbox will make this project public.  

![create project](img/new_create_project.png)  

After the project is created, you can browse repositories, users and logs using the navigation tab.  

![browse project](img/new_browse_project.png)  

All logs can be listed by clicking "Logs". You can apply a filter by username, or operations and dates under "Advanced Search".  

![browse project](img/new_project_log.png)  

##Managing members of a project 
###Adding members
You can add members with different roles to an existing project.  

![browse project](img/new_add_member.png)

###Updating and removing members
You can update or remove a member by clicking the icon on the right.  

![browse project](img/new_remove_update_member.png)

##Replicating images
Images replication is used to replicate repositories from one Harbor instance to another.  

The function is project-oriented, and once the system administrator set a policy to one project, all repositories under the project will be replicated to the remote registry. Each repository will start a job to run. If the project does not exist on the remote registry, a new project will be created automatically, but if it already exists and the user configured in policy has no write privilege to it, the process will fail. When a new repository is pushed to this project or an existing repository is deleted from this project, the same operation will also be replicated to the destination. The member information will not be replicated.  

There may be a bit of delay during replication according to the situation of the network. If replication job fails due to the network issue, the job will be re-scheduled a few minutes later.  

**Note:** The replication feature is incompatible between Harbor instance before version 0.3.5(included) and after version 0.3.5.  	

Start replication by creating a policy. Click "Add New Policy" on the "Replication" tab, fill the necessary fields, if there is no destination in the list, you need to create one, and then click "OK", a policy for this project will be created. If  "Enable" is chosen, the project will be replicated to the remote immediately.  

![browse project](img/new_create_policy.png)

You can enable, disable or delete a policy in the policy list view. Only policies which are disabled can be edited and only policies which are disabled and have no running jobs can be deleted. If a policy is disabled, the running jobs under it will be stopped.  

Click a policy, jobs which belong to this policy will be listed. A job represents the progress which will replicate a repository of one project to the remote.  

![browse project](img/new_policy_list.png)

##Searching projects and repositories
Entering a keyword in the search field at the top lists all matching projects and repositories. The search result includes both public and private repositories you have access privilege to.  

![browse project](img/new_search.png)

##Administrator options
###Managing user
Administrator can add "administrator" role to an ordinary user by toggling the switch under "Administrator". To delete a user, click on the recycle bin icon.  

![browse project](img/new_set_admin_remove_user.png)

###Managing destination
You can list, add, edit and delete destinations in the "Destination" tab. Only destinations which are not referenced by any policies can be edited.  

![browse project](img/new_manage_destination.png)

###Managing replication
You can list, edit, enable and disable policies in the "Replication" tab. Make sure the policy is disabled before you edit it.  

![browse project](img/new_manage_replication.png)

##Pulling and pushing images using Docker client

**NOTE: Harbor only supports Registry V2 API. You need to use Docker client 1.6.0 or higher.**  

Harbor supports HTTP by default and Docker client tries to connect to Harbor using HTTPS first, so if you encounter an error as below when you pull or push images, you need to add '--insecure-registry' option to /etc/default/docker (ubuntu) or /etc/sysconfig/docker (centos) and restart Docker:    
*FATA[0000] Error response from daemon: v1 ping attempt failed with error:  
Get https://myregistrydomain.com:5000/v1/_ping: tls: oversized record received with length 20527.   
If this private registry supports only HTTP or HTTPS with an unknown CA certificate,please add   
`--insecure-registry myregistrydomain.com:5000` to the daemon's arguments.  
In the case of HTTPS, if you have access to the registry's CA certificate, no need for the flag;  
simply place the CA certificate at /etc/docker/certs.d/myregistrydomain.com:5000/ca.crt*  

###Pulling images
If the project that the image belongs to is private, you should sign in first:  

```sh
$ docker login 10.117.169.182  
```
  
You can now pull the image:  

```sh
$ docker pull 10.117.169.182/library/ubuntu:14.04  
```

**Note: Replace "10.117.169.182" with the IP address or domain name of your Harbor node.**

###Pushing images
Before pushing an image, you must create a corresponding project on Harbor web UI. 

First, log in from Docker client:  

```sh
$ docker login 10.117.169.182  
```
  
Tag the image:  

```sh
$ docker tag ubuntu:14.04 10.117.169.182/demo/ubuntu:14.04  
``` 

Push the image:

```sh
$ docker push 10.117.169.182/demo/ubuntu:14.04  
```  

**Note: Replace "10.117.169.182" with the IP address or domain name of your Harbor node.**

##Deleting repositories

Repository deletion runs in two steps.  

First, delete a repository in Harbor's UI. This is soft deletion. You can delete the entire repository or just a tag of it. After the soft deletion, 
the repository is no longer managed in Harbor, however, the files of the repository still remain in Harbor's storage.  

![browse project](img/new_delete_repository.png)

**CAUTION: If both tag A and tag B refer to the same image, after deleting tag A, B will also get deleted.**  

Next, delete the actual files of the repository using the registry's garbage collection(GC). Make sure that no one is pushing images or Harbor is not running at all before you perform a GC. If someone were pushing an image while GC is running, there is a risk that the image's layers will be mistakenly deleted which results in a corrupted image. So before running GC, a preferred approach is to stop Harbor first.  

Run the below commands on the host which Harbor is deployed on to preview what files/images will be affected: 

```sh
$ docker-compose stop
$ docker run -it --name gc --rm --volumes-from registry registry:2.5.1 garbage-collect --dry-run /etc/registry/config.yml
```  
**NOTE:** The above option "--dry-run" will print the progress without removing any data.  

Verify the result of the above test, then use the below commands to perform garbage collection and restart Harbor. 

```sh
$ docker run -it --name gc --rm --volumes-from registry registry:2.5.1 garbage-collect  /etc/registry/config.yml
$ docker-compose start
```  

For more information about GC, please see [GC](https://github.com/docker/docker.github.io/blob/master/registry/garbage-collection.md).  
