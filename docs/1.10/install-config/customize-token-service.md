---
title: Customize the Harbor Token Service
---

By default, Harbor uses its own private key and certificate to authenticate with Docker clients. This topic describes how to optionally customize your configuration to use your own key and certificate.

Harbor requires Docker client to access the Harbor registry with a token. The procedure to generate a token is like [Docker Registry v2 authentication](https://github.com/docker/distribution/blob/master/docs/spec/auth/token.md). Firstly, you should make a request to the token service for a token. The token is signed by the private key. After that, you make a new request with the token to the Harbor registry, Harbor registry will verify the token with the public key in the rootcert bundle. Then Harbor registry will authorize the Docker client to push/pull images.

1. If you already have a certificate, go to step 3.
1. If not, you can generate a root certificate using openSSL with following commands:
  
    **1)Generate a private key:**

    ```sh
    openssl genrsa -out private_key.pem 4096
    ```
      
    **2)Generate a certificate:**

    ```sh
    openssl req -new -x509 -key private_key.pem -out root.crt -days 3650
    ```

You are about to be asked to enter information that will be incorporated into your certificate request. What you are about to enter is what is called a Distinguished Name or a DN. There are quite a few fields but you can leave some blank For some fields there will be a default valu. If you enter `.`, the field will be left blank. Following are what you're asked to enter.

Country Name (2 letter code) [AU]:

State or Province Name (full name) [Some-State]:

Locality Name (eg, city) []:

Organization Name (eg, company) [Internet Widgits Pty Ltd]:

Organizational Unit Name (eg, section) []:

Common Name (eg,  server FQDN or YOUR name) []:

Email Address []:

After you execute these two commands, you will see private_key.pem and root.crt in the **current directory**, just type "ls", you'll see them.

3. Refer to [Installation Guide](../installation-guide.md) to install Harbor, After you execute ./prepare, Harbor generates several config files. We need to replace the original private key and certificate with your own key and certificate.

4.Replace the default key and certificate. Assume that you key and certificate are in the directory /root/cert, following are what you should do:

    ```shell
    cd config/ui
    cp /root/cert/private_key.pem private_key.pem
    cp /root/cert/root.crt ../registry/root.crt
    ```

5.After these, go back to the make directory, you can start Harbor using following command:

    ```shell
    docker-compose up -d
    ```

6.Then you can push/pull images to see if your own certificate works.
