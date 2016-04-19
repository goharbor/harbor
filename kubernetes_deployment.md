## Deploy harbor on kubernetes.
For now, it's a little tricky to start harbor on kubernetes because
  1. Registry uses https, so we need cert or workaround to avoid errors like this:
     ```
     Error response from daemon: invalid registry endpoint https://{HOST}/v0/: unable to ping registry endpoint https://{HOST}/v0/
     v2 ping attempt failed with error: Get https://{HOST}/v2/: EOF
     v1 ping attempt failed with error: Get https://{HOST}/v1/_ping: EOF. If this private registry supports only HTTP or HTTPS with an unknown CA certificate, please add `--insecure-registry {HOST}` to the daemon's arguments. In the case of HTTPS, if you have access to the registry's CA certificate, no need for the flag; simply place the CA certificate at /etc/docker/certs.d/{HOST}/ca.crt
     ```
     There is a workaround if you don't have a cert. The workaround is to add the host into the list of insecure registry by editting the ```/etc/default/docker``` file:
     ```
     sudo vi /etc/default/docker
     ```
     add the line at the end of file:
     ```
     DOCKER_OPTS="$DOCKER_OPTS --insecure-registry={HOST}"
     ```
     restart docker service
     ```
     sudo service docker restart
     ```

  2. The registry config file need to know the IP (or DNS name) of the registry, but on kubernetes, you won't know the IP before the service is created. There are several workarounds to solve this problem for now:
     - Use DNS name and link th DNS name with the IP after the service is created.
     - Rebuild the registry image with the service IP after the service is created and use ```kubectl rolling-update``` to update to the new image.
        
 
To start harbor on kubernetes, you first need to build the docker images. The docker images for deploying Harbor on Kubernetes depends on the docker images to deploy Harbor with docker-compose. So the first step is to build docker images with docker-compose. Before actually building the images, you need to first adjust the [configuration](https://github.com/vmware/harbor/blob/master/Deploy/harbor.cfg):
- Change the [hostname](https://github.com/vmware/harbor/blob/master/Deploy/harbor.cfg#L5) to ```localhost```
- Adjust the [email settings](https://github.com/vmware/harbor/blob/master/Deploy/harbor.cfg#L11) according to your needs.

Then you can run the following commends to build docker images:
```
cd Deploy
./prepare
docker-compose build
docker build -f kubernetes/dockerfiles/proxy-dockerfile -t {your_account}/proxy .
docker build -f kubernetes/dockerfiles/registry-dockerfile -t {your_account}/registry .
docker build -f kubernetes/dockerfiles/ui-dockerfile -t {your_account}/deploy_ui .
docker tag deploy_mysql {your_account}/deploy_mysql
docker push {your_account}/proxy
docker push {your_account}/registry
docker push {your_account}/deploy_ui
docker push {your_account}/deploy_mysql
```
  
where "your_account" is your own registry. Then you need to update the "image" field in the ```*-rc.yaml``` files at:
```
Deploy/kubernetes/mysql-rc.yaml
Deploy/kubernetes/proxy-rc.yaml
Deploy/kubernetes/registry-rc.yaml
Deploy/kubernetes/ui-rc.yaml
```

Further more, the following configuration could be changed according to your need:
 - **harbor_admin_password**: The password for the administrator of Harbor, by default the password is Harbor12345. You can changed it [here](https://github.com/vmware/harbor/blob/master/Deploy/kubernetes/ui-rc.yaml#L36).
 - **auth_mode**: The authentication mode of Harbor. By default it is *db_auth*, i.e. the credentials are stored in a database. Please set it to *ldap_auth* if you want to verify user's credentials against an LDAP server.  You can change the configuration [here](https://github.com/vmware/harbor/blob/master/Deploy/kubernetes/ui-rc.yaml#L40).
 - **ldap_url**: The URL for LDAP endpoint, for example ldaps://ldap.mydomain.com. It is only used when **auth_mode** is set to *ldap_auth*.  It could be changed [here](https://github.com/vmware/harbor/blob/master/Deploy/kubernetes/ui-rc.yaml#L42).
 - **ldap_basedn**: The basedn template for verifying the user's credentials against LDAP, for example uid=%s,ou=people,dc=mydomain,dc=com.  It is only used when **auth_mode** is set to *ldap_auth*.  It could be changed [here](https://github.com/vmware/harbor/blob/master/Deploy/kubernetes/ui-rc.yaml#L44).
 - **db_password**: The password of root user of mySQL database. Change this password for any production use.  You need to change both [here](https://github.com/vmware/harbor/blob/master/Deploy/kubernetes/ui-rc.yaml#L28) and [here](https://github.com/vmware/harbor/blob/master/Deploy/harbor.cfg#L32) to make the change. Please note, you need to change the ```harbor.cfg``` before building the docker images.

Finally you can start the jobs by running:
```
kubectl create -f Deploy/kubernetes
```

