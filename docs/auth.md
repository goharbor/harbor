#Customize Harbor auth with your key and certificate

By default, Harbor use default private key and certificate in authentication.  The auth procedure is like [Docker Registry v2 authentication](https://github.com/docker/distribution/blob/master/docs/spec/auth/token.md). Also, you can customize your configuration with your own key and certificate with the following steps:

1.If you already have a certificate, go to step 3.

2.If not, you can generate a self-signed certificate using openSSL with following commands
  
**1)Generate a private key:**


```sh
    openssl genrsa -out private_key.pem 2048    
```

you can call it prvtkey.pem or other names you like.
    
   
**2)Generate a certificate:**

```sh
    openssl req -new -x509 -key private_key.pem -out root.crt -days 1095
```    
   
3.Refer to [Installation Guide](https://github.com/vmware/harbor/blob/master/docs/installation_guide.md) to install Harbor, After you execute ./prepare, Harbor generates several config files. We need to replace the original private key and certificate with your own key and certificate.

4.Following are what you should do:
 
**1)cd config/ui, you will see private_key.pem.**
    
**2)replace private_key.pem with your private_key.pem**
    
**4)cd ../registry, you will see root.crt. Replace it with your certificate root.crt**
 

5.After these, go back to the Deploy directory, you can start Harbor using following command:
```
  docker-compose up -d
```

6.Then you can push/pull images to see if your own certificate works. Please refer [User Guide](https://github.com/vmware/harbor/blob/master/docs/user_guide.md)


