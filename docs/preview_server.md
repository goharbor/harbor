**Important!** 
 - Please note that this preview server is **ONLY** for your experience Harbor purpose. 
 - Please **DO NOT** upload any sensitive images to this server. 
 - We will **CLEAN AND RESET** the server every **TWO Days**.
 - You can only experience the none-admin functionalities on this server. Please follow the **[Installation Guide](installation_guide.md)** to setup a Harbor server locally to try more advanced features.
 - Please do not push large images(>100MB) as the server has limited storage.

If you encounter any questions during use the preview server please contact us at harbor@ vmware.com

**Usage**

 - 1> The address of the preview server is [https://ec2-52-14-7-203.us-east-2.compute.amazonaws.com](https://ec2-52-14-7-203.us-east-2.compute.amazonaws.com)
 - 2> You can registry a new user by yourself.
 - 3> As this preview server use self-signed certificate, if you want to use docker client to talk with this server. You need to save the CA certificate to you local machine:
     - Create a folder ```/etc/docker/certs.d/ec2-52-14-7-203.us-east-2.compute.amazonaws.com``` on you local host.  
     - Save the [ca.crt](ca.crt) to ```/etc/docker/certs.d/ec2-52-14-7-203.us-east-2.compute.amazonaws.com/ca.crt``` to your local host.
 
 - 4> Then you can use the account/password created in step 2 to login 
 ```
 docker login ec2-52-14-7-203.us-east-2.compute.amazonaws.com
 ```
You can refer to [User Guide](user_guide.md) for more details on how to use Harbor.
