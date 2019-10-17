# Other General Settings

## Managing self-registration
You can manage whether a user can sign up for a new account. This option is not available if you use LDAP authentication.  
![browse project](../img/new_self_reg.png)

## Managing email settings
You can change Harbor's email settings, the mail server is used to send out responses to users who request to reset their password.  
![browse project](../img/new_config_email.png)

## Managing registry read only
You can change Harbor's registry read only settings, read only mode will allow 'docker pull' while preventing 'docker push' and the deletion of repository and tag.
![browse project](../img/read_only.png)

If it set to true, deleting repository, tag and pushing image will be disabled. 
![browse project](../img/read_only_enable.png)


```
$ docker push 10.117.169.182/demo/ubuntu:14.04  
The push refers to a repository [10.117.169.182/demo/ubuntu]
0271b8eebde3: Preparing 
denied: The system is in read only mode. Any modification is prohibited.  
```