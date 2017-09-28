
## Integration with Kubernetes
This Document decribes how to deploy Harbor on Kubernetes.  It has been verified on **Kubernetes v1.6.5** and **Harbor v1.2.0**

### Prerequisite
* You need to download docker images of Harbor. 
	* Download the offline installer of Harbor v1.2.0 from the [release](https://github.com/vmware/harbor/releases) page.
	* Uncompress the offline installer and get the images tgz file harbor.*.tgz.
	* Load the images into docker:  
		```
		docker load -i harbor.*.tgz 
		```
* You should have domain knowledge about Kubernetes (Replication Controller, Service, Persistent Volume, Persistent Volume Claim, Config Map). 

### Configuration
We provide a python script `make/kubernetes/prepare` to generate Kubernetes ConfigMap files. 
The script is written in python, so you need a version of python in your deployment environment.
Also the script need `openssl` to generate private key and certification, make sure you have a workable `openssl`. 

There are some args of the python script:

- -f: Default Value is `../harbor.cfg`. You can specify other config file of Harbor.
- -k: Path to https private key. This arg can overwrite the value of `ssl_cert_key` in `harbor.cfg`.
- -c: Path to https certification. This arg can overwrite the value of `ssl_cert` in `harbor.cfg`.
- -s: Path to secret key. Must be 16 characters. If you don't set it, the script will generate it automatically. 

#### Basic Configuration
These Basic Configuration must be set. Otherwise you can't deploy Harbor on Kubernetes.

- `make/harbor.cfg`: Basic config of Harbor. Please refer to `harbor.cfg`.

  ```
  #Hostname is the endpoint for accessing Harbor,
  #To accept access from outside of Kubernetes cluster, it should be set to a worker node.
  hostname = 10.192.168.5
  ```
- `make/kubernetes/**/*.svc.yaml`: Specify the service of pods.  In particular, the externalIP should be set in `make/kubernetes/nginx/nginx.svc.yaml`:

  ```
  ...
  metadata:
      name: nginx
  spec:
      ports:
      - name: http
        port: 80
      selector:
        name: nginx-apps
      externalIPs:
        - 10.192.168.5
  ``` 
  
- `make/kubernetes/**/*.rc.yaml`: Specify configs of containers.  
- `make/kubernetes/pv/*.pvc.yaml`: Persistent Volume Claim.  
  You can set capacity of storage in these files. example:

  ```
  resources:
    requests:
      # you can set another value to adapt to your needs
      storage: 100Gi
  ```

- `make/kubernetes/pv/*.pv.yaml`: Persistent Volume. Be bound with `*.pvc.yaml`.  
  PVs and PVCs are one to one correspondence. If you changed capacity of PVC, you need to set capacity of PV together.
  example:

  ```
  capacity:
    # same value with PVC
    storage: 100Gi
  ```

  In PV, you should set another way to store data rather than `hostPath`:

  ```
  # it's default value, you should use others like nfs.
  hostPath:
    path: /data/registry
  ```

  For more infomation about storage solution, Please check [Kubernetes Document](http://kubernetes.io/docs/user-guide/persistent-volumes/) 

Then you can generate ConfigMap files by :

```
python make/kubernetes/prepare
```

These files will be generated:

- make/kubernetes/jobservice/jobservice.cm.yaml
- make/kubernetes/mysql/mysql.cm.yaml
- make/kubernetes/nginx/nginx.cm.yaml
- make/kubernetes/registry/registry.cm.yaml
- make/kubernetes/ui/ui.cm.yaml
- make/kubernetes/adminserver/adminserver.cm.yaml

#### Advanced Configuration
If Basic Configuration was not covering your requirements, you can read this section for more details.

`./prepare` has a specify format of placeholder:

- `{{key}}`: It means we should replace the placeholder with the value in `config.cfg` which name is `key`.
- `{{num key}}`: It's used for multiple lines text. It will add `num` spaces to the leading of every line in text.

You can find all configs of Harbor in `make/kubernetes/templates/`. There are specifications of these files:

- `jobservice.cm.yaml`: ENV and web config of jobservice
- `mysql.cm.yaml`: Root passowrd of MySQL
- `nginx.cm.yaml`: Https certification and nginx config. If you are fimiliar with nginx, you can modify it. 
- `registry.cm.yaml`: Token service certification and registry config
  Registry use filesystem to store data of images. You can find it like:

  ```
  storage:
      filesystem:
        rootdirectory: /storage
  ``` 

  If you want use another storage backend, please see [Docker Doc](https://docs.docker.com/datacenter/dtr/2.1/guides/configure/configure-storage/)
- `ui.cm.yaml`: Token service private key, ENV and web config of ui.
- `adminserver.cm.yaml`: Initial values of configuration attributes of Harbor.

`ui`, `jobservice` and `adminserver` are powered by beego. If you are fimiliar with beego, you can modify configs in `ui.cm.yaml`, `jobservice.cm.yaml` and `adminserver.cm.yaml`.


### Running
When you finished your configuring and generated ConfigMap files, you can run Harbor on kubernetes with these commands:

```
# create pv & pvc
kubectl apply -f make/kubernetes/pv/log.pv.yaml
kubectl apply -f make/kubernetes/pv/registry.pv.yaml
kubectl apply -f make/kubernetes/pv/storage.pv.yaml
kubectl apply -f make/kubernetes/pv/log.pvc.yaml
kubectl apply -f make/kubernetes/pv/registry.pvc.yaml
kubectl apply -f make/kubernetes/pv/storage.pvc.yaml

# create config map
kubectl apply -f make/kubernetes/jobservice/jobservice.cm.yaml
kubectl apply -f make/kubernetes/mysql/mysql.cm.yaml
kubectl apply -f make/kubernetes/nginx/nginx.cm.yaml
kubectl apply -f make/kubernetes/registry/registry.cm.yaml
kubectl apply -f make/kubernetes/ui/ui.cm.yaml
kubectl apply -f make/kubernetes/adminserver/adminserver.cm.yaml

# create service
kubectl apply -f make/kubernetes/jobservice/jobservice.svc.yaml
kubectl apply -f make/kubernetes/mysql/mysql.svc.yaml
kubectl apply -f make/kubernetes/nginx/nginx.svc.yaml
kubectl apply -f make/kubernetes/registry/registry.svc.yaml
kubectl apply -f make/kubernetes/ui/ui.svc.yaml
kubectl apply -f make/kubernetes/adminserver/adminserver.svc.yaml

# create k8s rc
kubectl apply -f make/kubernetes/registry/registry.rc.yaml
kubectl apply -f make/kubernetes/mysql/mysql.rc.yaml
kubectl apply -f make/kubernetes/jobservice/jobservice.rc.yaml
kubectl apply -f make/kubernetes/ui/ui.rc.yaml
kubectl apply -f make/kubernetes/nginx/nginx.rc.yaml
kubectl apply -f make/kubernetes/adminserver/adminserver.rc.yaml
```

After the pods are running, you can access Harbor's UI via the configured endpoint `10.192.168.5` or issue docker commands such as `docker login 10.192.168.5` to interact with the registry.
