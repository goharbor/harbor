## Make use of pre-built images of Harbor

Community members have helped building Harbor's docker images. If you want to save time from building Harbor from source, please follow the below instructions to quickly pull Harbor's pre-built images for installation. 

### Steps

Run the command `update_compose.sh` :
```
$ ./update_compose.sh 
 
Please enter the registry service you want to pull the pre-built images from.
Enter 1 for Docker Hub.
Enter 2 for Daocloud.io (recommended for Chinese users).
or enter other registry URL such as https://my_registry/harbor/ .
The default is 1 (Docker Hub): 
```

Enter **1** to pull images from Docker Hub,  
Enter **2** to pull image from Daocloud.io, recommended for Chinese users.  
or Enter other registry URL like `https://my_registry/harbor/` . Do not forget the "/" and the end.

This command backs up and updates the file `Deploy/docker-compose.yml` . Next, just follow the [Harbor Installation Guide](../../docs/installation_guide.md) to install Harbor. 

