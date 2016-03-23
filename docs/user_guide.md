#User Guide
##Overview
This guide takes you through the fundamentals of using Harbor. You'll learn how to use Harbor to:  

* Manage your projects.
* Manage members of a project.
* Search projects and repositories.
* Manage Harbor system if you are the system administrator.
* Pull and push images using Docker client.


##Role Based Access Control
RBAC (Role Based Access Control) is provided in Harbor and there are four roles with different privileges:  

* **Guest**: Guest has read-only privilege for a specified project.
* **Developer**: Developer has read and write privileges for a project.
* **ProjectAdmin**: When creating a new project, you will be assigned the "ProjectAdmin" role to the project. Besides read-write privileges, the "ProjectAdmin" also has some management privileges, such as adding and removing members.
* **SysAdmin**: "SysAdmin" has the most privileges. In addition to the privileges mentioned above, "SysAdmin" can also list all projects, set an ordinary user as administrator and delete users. The public project "library" is also owned by the administrator.  
* **Anonymous**: When a user is not logged in, the user is considered as an "anonymous" user. An anonymous user has no access to private projects and has read-only access to public projects.  

##User account
As a new user, you can sign up an account by going through the self-registration process. The username and email must be unique in the Harbor system. The password must contain at least 7 characters with 1 lowercase letter, 1 uppercase letter and 1 numeric character.  

If the administrator has configured LDAP/AD as authentication source, no sign-up is required. The LDAP/AD user id can be used directly to log in to Harbor.  
  
When you forgot your password, you can follow the below steps to reset the password:  

1. Click the link "forgot password" in the sign in page.
2. Input the email used when you signed up, an email will be sent out to you.
3. After receiving the email, click on the link in the email which directs you to a password reset web page.
4. Input your new password and click "Submit".


##Managing projects
A project in Harbor contains all repositories of an application. RBAC is applied to a project. There are two types of projects in Harbor:  

* **Public**: All users have the read privilege to a public project, it's convenient for you to share some repositories with others in this way.
* **Private**: A private project can only be accessed by users with proper privileges.  

You can create a project after you signed in. Enabling the "Public project" checkbox will make this project public.  

![create project](img/create_project.png)  

After the project is created, you can browse repositories, users and access logs using the navigation column on the left.  

![browse project](img/browse_project.png)  

All access logs can be listed by clicking "Logs". You can apply a filter by username, or operations and dates under "Advanced Search".  

![browse project](img/project_log.png)  

##Managing members of a project 
###Adding members
You can add members with different roles to an existing project.  

![browse project](img/add_member.png)

###Updating and removing members
You can update or remove a member by clicking the icon on the right.  

![browse project](img/remove_update_member.png)

##Searching projects and repositories
Entering a keyword in the search field at the top lists all matching projects and repos. The search result includes public repos and private repos you have access privilege to.  

![browse project](img/search.png)

##Administrator options
###Setting administrator and deleting user
Administrator can add "SysAdmin" role to an ordinary user by toggling the switch under "System Admin". To delete a user, click on the recycle bin icon.  

![browse project](img/set_admin_remove_user.png)

##Pulling and pushing images using Docker client

**NOTE: Harbor only supports Registry V2 API. You need to use Docker client 1.6.0 or higher.**  

Harbor supports HTTP by default and Docker client trys to connect to Harbor using HTTPS first, so if you encounter an error as below when you pull or push images, you need to add '--insecure-registry' option to /etc/default/docker (ubuntu) or /etc/sysconfig/docker (centos):    
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