[Back to table of contents](../_index.md)

----------

# Configure Global Settings

You can configure Harbor to connect to an email server, and set the registry in read-only mode.

## Configure an Email Server

You can change Harbor's email settings, the mail server is used to send out responses to users who request to reset their password.  
![browse project](../../img//new_config_email.png)

## Make the Registry Read Only

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

----------

[Back to table of contents](../_index.md)