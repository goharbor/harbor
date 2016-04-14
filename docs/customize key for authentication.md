#Customize Harbor auth with your key and certificate

Harbor requires Docker client to access the Harbor registry with a token. The procedure to generate a token is like [Docker Registry v2 authentication](https://github.com/docker/distribution/blob/master/docs/spec/auth/token.md). Firstly, you should make a request to the token service for a token. The token is signed by the private key. After that, you make a new request with the token to the Harbor registry, Harbor registry will verify the token with the public key in the rootcert bundle. Then Harbor registry will authorize the Docker client to push/pull images.

By default, Harbor uses default private key and certificate in authentication. Also, you can customize your configuration with your own key and certificate with the following steps:

1.If you already have a certificate, go to step 3.

2.If not, you can generate a root certificate using openSSL with following commands:
  
**1)Generate a private key:**


```sh
    openssl genrsa -out private_key.pem 4096    
```
   
**2)Generate a certificate:** 

```sh
    openssl req -new -x509 -key private_key.pem -out root.crt -days 3650
```    
   
3.Refer to [Installation Guide](https://github.com/vmware/harbor/blob/master/docs/installation_guide.md) to install Harbor, After you execute ./prepare, Harbor generates several config files. We need to replace the original private key and certificate with your own key and certificate.

4.Following are what you should do:
 
**1)cd config/ui, you will see private_key.pem.**
    
**2)replace private_key.pem with your private_key.pem**
    
**3)cd ../registry, you will see root.crt. Replace it with your root.crt**
 

5.After these, go back to the Deploy directory, you can start Harbor using following command:
```
  docker-compose up -d
```

6.Then you can push/pull images to see if your own certificate works. Please refer [User Guide](https://github.com/vmware/harbor/blob/master/docs/user_guide.md) for more info.


