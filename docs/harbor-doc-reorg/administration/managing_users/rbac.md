# Harbor Role Based Access Control (RBAC)  

![rbac](../../img/rbac.png)

Harbor manages images through projects. Users can be added into one project as a member with one of the following different roles:  

* **Limited Guest**: A Limited Guest does not have full read privileges for a project. They can pull images but cannot push, and they cannot see logs or the other members of a project. For example, you can create limited guests for users from different organizations who share access to a project.
* **Guest**: Guest has read-only privilege for a specified project. They can pull and retag images, but cannot push.
* **Developer**: Developer has read and write privileges for a project.
* **Master**: Master has elevated permissions beyond those of 'Developer' including the ability to scan images, view replications jobs, and delete images and helm charts. 
* **ProjectAdmin**: When creating a new project, you will be assigned the "ProjectAdmin" role to the project. Besides read-write privileges, the "ProjectAdmin" also has some management privileges, such as adding and removing members, starting a vulnerability scan.

Besides the above roles, there are two system-level roles:  

* **Harbor system administrator**: "Harbor system administrator" has the most privileges. In addition to the privileges mentioned above, "Harbor system administrator" can also list all projects, set an ordinary user as administrator, delete users and set vulnerability scan policy for all images. The public project "library" is also owned by the administrator.  
* **Anonymous**: When a user is not logged in, the user is considered as an "Anonymous" user. An anonymous user has no access to private projects and has read-only access to public projects.  

For full details of the permissions of the different roles, see [User Permissions By Role](user_permissions_by_role.md).

[Configure Harbor User Settings at the Command Line](configure_user_settings_cli.md)

## Create User Accounts
	
In database authentication mode, the Harbor system administrator creates user accounts manually. 

1. Log in to the Harbor interface with an account that has Harbor system administrator privileges.
1. Under **Administration**, go to **Users**.

   ![Create user account](../../img/create_user.png)
1. Click **New User**.
1. Enter information about the new user.

   ![Provide user information](../../img/new_user.png)

   - The username must be unique in the Harbor system
   - The email address is used for password recovery
   - The password must contain at least 8 characters with 1 lowercase letter, 1 uppercase letter and 1 numeric character

If users forget their password, there is a **Forgot Password** in the Harbor log in page.

## Assigning the Administrator Role

Harbor system administrators can assign the Harbor system administrator role to other users by selecting usernames and clicking Set as Administrator in the **Users** tab. 

![browse project](../../img/new_set_admin_remove_user.png)

To delete users, select a user and click `DELETE`. Deleting user is only supported under database authentication mode.