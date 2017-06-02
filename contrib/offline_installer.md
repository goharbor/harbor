# Harbor offline installer
## Overview
This guide takes you through the installation of Harbor using an offline installer. The installer will install docker(1.10.0+) and docker-compose(1.6.0+) first if they don't exist on the host where Harbor will be deployed. And if there is no images in your environment, you can also download the prepared busybox image and load it to Harbor.

##Prerequisites
The installer contains a script used to install docker which only works on Ubuntu 14.04, so if you want to install Harbor on other Linux distribution, please install docker(1.10.0+) by yourself first.  

Linux Distribution | Docker | Support
------------ | ------------- | -------------
Ubuntu 14.04 |  Not Required | Yes
Other Linux Distribution |  Docker(1.10.0+) Required | Yes
  
## Install Harbor
1.Download the [harbor-0.3.0-installer.tar](http://bintray.com/xxx/xxx.tar).  
2.Decompress it:
```sh
tar -xvf harbor-0.3.0-installer.tar
```  
3.Run the installer:
```sh
harbor/install.sh -h 192.168.0.2
```

Replace 192.168.0.2 with your IP address or hostname which is used to access admin UI and registry service. DO NOT use localhost or 127.0.0.1, because Harbor needs to be accessed by external clients.

**Notes:**At the very least, you will just need provide the -h(--host) option to run the installer. If you need more configuretions, you can edit the harbor.cfg under directory harbor/Deploy. If you have configured the hostname attribute in the harbor.cfg, the -h(--host) option is not necessary. About more details, please see the [installation guide](https://github.com/vmware/harbor/blob/master/docs/installation_guide.md).  

## Load prepared image
1.Download the [busybox image](https://bintray.com/harbor/generic/download_file?file_path=busybox.tar).  
2.Load it to Harbor:
```sh
docker load -i busybox.tar 
```