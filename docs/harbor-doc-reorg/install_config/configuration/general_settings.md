# Administrator Options

## Managing Project Creation
Use the **Project Creation** drop-down menu to set which users can create projects. Select **Everyone** to allow all users to create projects. Select **Admin Only** to allow only users with the Administrator role to create projects.  
![browse project](../../img//new_proj_create.png)

## Managing Email Settings
You can change Harbor's email settings, the mail server is used to send out responses to users who request to reset their password.  
![browse project](../../img//new_config_email.png)

## Managing Registry Read Only
You can change Harbor's registry read only settings, read only mode will allow 'docker pull' while preventing 'docker push' and the deletion of repository and tag.
![browse project](../../img//read_only.png)

If it set to true, deleting repository, tag and pushing image will be disabled. 
![browse project](../../img//read_only_enable.png)


```
$ docker push 10.117.169.182/demo/ubuntu:14.04  
The push refers to a repository [10.117.169.182/demo/ubuntu]
0271b8eebde3: Preparing 
denied: The system is in read only mode. Any modification is prohibited.  
```