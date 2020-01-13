[Back to table of contents](../index.md)

----------

# Make the Registry Read Only

You can set Harbor to read-only mode. In read-only mode, Harbor allows `docker pull` but prevents `docker push` and the deletion of repositories and tags.

![Read-only mode](../../img//read_only.png)

If it set to true, deleting repositories, tags and pushing images are not permitted.

![browse project](../../img//read_only_enable.png)


```
$ docker push 10.117.169.182/demo/ubuntu:14.04  
The push refers to a repository [10.117.169.182/demo/ubuntu]
0271b8eebde3: Preparing 
denied: The system is in read only mode. Any modification is prohibited.  
```

----------

[Back to table of contents](../index.md)