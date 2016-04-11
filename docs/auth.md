#Customize harbor auth with your key and certificate

By default, harbor use default private key and certificate in authentication.  The auth procedure is like [Docker Registry v2 authentication](https://github.com/docker/distribution/blob/master/docs/spec/auth/token.md). Also, you can customize your configuration with your own key and certificate with the following steps:

1.If you already have a certificate, go to step 3.

2.If not, you can generate a self-signed certificate using openSSL with following commands
  
**1)Generate a private key:**


```sh
    openssl genrsa -out prvtkey.pem 2048    
```

you can call it prvtkey.pem or other names you like.
    
   
**2)Generate a certificate:**

```sh
    openssl req -new -x509 -key prvtkey.pem -out cacert.pem -days 1095
```    
   
prvtkey.pem is what you generated in the first step, if you change the name, you should change it in the command. Also you can name cacert.pem what you like.

3.Refer to [Installation Guide](https://github.com/vmware/harbor/blob/master/docs/installation_guide.md) to install harbor. After you execute ./prepare, harbor generates several config files. We need to replace the original private key and certificate with your own key and certificate.

4.Following are what you should do:
 
**1)edit docker-compose.yml, find private_key.pem replace it with your own private key as following:**


![edit docker-compose.yml](img/edit_docker-compose-yml.png)

![edit docker-compose.yml](img/after_edit_docker-compose-yml.png)

**2)cd config/ui, you will see private_key.pem.**
    
**3)replace private_key.pem with your private key.**
    
**4)cd ../registry, you will see root.crt. Replace it with your certificate.**
 
**5)at the same directory, you will see config.yml. We need to modify it, open it and find root.crt, then change it to your certificate.**

5.After these, go back to harbor directory, execute:

```sh
       docker-compose build
```
```sh
       docker-compose up –d  
```

6.Then you can push/pull images to see if your own certificate works. Please refer [User Guide](https://github.com/vmware/harbor/blob/master/docs/user_guide.md)


