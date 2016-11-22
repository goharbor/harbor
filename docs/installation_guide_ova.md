# Installing and Configuring Harbor on vSphere as Virtual Appliance

This guide walks you through the steps about installing and configuring Harbor on vSphere as an virtual appliance (OVA). If you are installing Harbor on a Linux host, refer to this **[Installation Guide](installation_guide.md)**.

## Prerequisites
* vCenter 5.5+ and at least an ESX host. 
* 2 vCPUs, 4GB memory and 80GB free disk space in datastore.
* A network with DHCP capability, or a static IP address for the virtual appliance.

## Planning for installation

### User management  
By default, Harbor stores user information in an internal database. Harbor can also be configured to authenticate against an external LDAP or AD server. The proper **authentication mode** must be set at the deployment time. 

**NOTE: This mode cannot be changed after the first boot of Harbor.**

### Security 

By default, Harbor uses HTTPS for secure communication. A self-signed certificate is generated at first boot. A Docker client or a VCH (Virtual Container Host) needs to trust Harbor's CA certificate in order to interact with Harbor. 

The self-generated certificate can be replaced by supplying a certificate signed by other CAs in OVA's settings. 

Harbor can be configured to use plain HTTP for  some environments like testing or continuous integration (CI). However, it is **NOT** recommended to use HTTP for production because the communication is never secure.

### Networking

Harbor can obtain IP address by DHCP. This is convenient for testing purpose. For a production system, it is recommended that static IP address be used.


## Installation
1. Download the OVA file to your local disk from the **[official release page](https://github.com/vmware/harbor/releases)**.  

2. Log in vSphere web client. Right click on the datacenter, cluster or host which Harbor will be deployed on. Select "Deploy OVF Template" and open the import wizard.  

 ![ova](img/ova/ova01.png)

3. Select the OVA file from your local disk and click "Next".  

 ![ova](img/ova/ova02.png)

4. Review the OVF template details and click "Next".  

 ![ova](img/ova/ova03.png)

5. Accept the end user license agreements and click "Next".  

 ![ova](img/ova/ova04.png)

6. Specify a name and a location for the virtual appliance.  

 ![ova](img/ova/ova05.png)

7. Select the datastore and virtual disk format, click "Next".  

 ![ova](img/ova/ova06.png)

8. Configure the network(s) the virtual appliance should be connected to.  

 ![ova](img/ova/ova07.png)

9. Customize the properties of Harbor. The properties are described below. Note that at the very least, you just need to set four properties: **Root Password**, **Harbor Admin Password**,  **Database Password** and **Authentication Mode**.  

 ![ova](img/ova/ova08.png)

 * System
	* **Root Password**: The initial password of the root user. Subsequent changes of password should be performed in operating system. (8-128 characters)
	* **Harbor Admin Password**: The initial password of Harbor admin. It only works for the first time when Harbor starts. It has no effect after the first launch of Harbor. Change the admin password from UI after launching Harbor. 
	* **Database Password**: The initial password of the root user of MySQL database. Subsequent changes of password should be performed in operating system. (8-128 characters)
	* **Permit Root Login**: Specifies whether root user can log in using SSH.
	* **Self Registration**: Determine whether the self-registration is allowed or not. Set this to off to disable a user's self-registration in Harbor. This flag has no effect when users are stored in LDAP or AD.
	* **Garbage Collection**: When setting this to true, Harbor performs garbage collection everytime it boots up. The first time setting this flag to true needs to power off the VM and power it on again.

 * Authentication

    The **Authentication Mode** must be set before the first boot of Harbor. Subsequent changes to **Authentication Mode** does not have any effect. When **ldap_auth** mode is enabled, properties related to LDAP/AD must be set.
  
	* **Authentication Mode**: The default authentication mode is **db_auth**. Set it to **ldap_auth** when users' credentials are stored in an LDAP or AD server. Note: this option can only be set once.
	* **LDAP URL**: The URL of an LDAP/AD server.
	* **LDAP Search DN**: A user's DN who has the permission to search the LDAP/AD server. Leave blank if your LDAP/AD server supports anonymous search, otherwise you should configure this DN and **LDAP Seach Password**.
	* **LDAP Search Password**: The password of the user for LDAP search. Leave blank if your LDAP/AD server supports anonymous search.
	* **LDAP Base DN**: The base DN of a node from which to look up a user for authentication. The search scope includes subtree of the node.
	* **LDAP UID**: The attribute used in a search to match a user, it could be uid, cn, email, sAMAccountName or other attributes depending on your LDAP/AD server.

 * Security
 
    If HTTPS is enabled, a self-signed certificate is generated by default. To supply your own certificate, please fill in **SSL Cert** and **SSL Cert Key**. Do not use HTTP in any production system.
 
	* **Protocol**: The protocol for accessing Harbor. Warning: setting it to http makes the communication insecure.
	* **SSL Cert**: Paste in the content of a certificate file. Leave blank for a generated self-signed certificate.
	* **SSL Cert Key**: Paste in the content of a certificate key file. Leave blank for a generated key.
	* **Verify Remote Cert**: Determine whether the image replication should verify the certificate of a remote Harbor registry. Set this flag to off when the remote registry uses a self-signed or untrusted certificate.

 * Email Settings
 
   To allow a user to reset his/her own password through email, configure the below email settings:
 
	* **Email Server**: The mail server to send out emails to reset password. 
	* **Email Server Port**: The port of mail server.
	* **Email Username**: The user from whom the password reset email is sent.
	* **Email Password**: The password of the user from whom the password reset email is sent.
	* **Email From**: The name of the email sender.
	* **Email SSL**: Whether to enabled secure mail transmission.

 * Networking properties
	* **Default Gateway**: The default gateway address for this VM. Leave blank if DHCP is desired.
	* **Domain Name**: The domain name of this VM. Leave blank if DHCP is desired.
	* **Domain Search Path**: The domain search path(comma or space separated domain names) for this VM. Leave blank if DHCP is desired.
	* **Domain Name Servers**: The domain name server IP Address for this VM(comma separated). Leave blank if DHCP is desired.
	* **Network 1 IP Address**: The IP address of this interface. Leave blank if DHCP is desired.
	* **Network 1 Netmask**: The netmask or prefix for this interface. Leave blank if DHCP is desired.

 **Notes:** If you want to enable HTTPS with a self-signed certificate created manually, refer to the "Getting a certificate" part of this [guide](https://github.com/vmware/harbor/blob/master/docs/configure_https.md#getting-a-certificate) for generating a certificate.  

 After you complete the properties, click "Next".  

10. Review your settings and click "Finish" to complete the deployment.  

 ![ova](img/ova/ova09.png)

11. Power on the virtual appliance. It may take a few minutes for the first bootup. The virtual appliance needs to initialize itself for configuration like netowrk address and password. 

12. When the appliance is ready, check from vSphere Web Client for its IP address. Open a browser and type in the URL `http(s)://harbor_ip_address` or `http(s)://harbor_host_name`. Log in as the admin user and verify Harbor has been successfully installed. 

13. For information on how to use Harbor, please refer to [User Guide of Harbor](user_guide.md).

## Getting Certificate of Harbor's CA

By default, Harbor uses a self-signed certificate in HTTPS. A Docker client or a VCH needs to trust Harbor's CA certificate in order to interact with Harbor. 
To download Harbor's CA certificate and import into a Docker client, follow the below steps:

1. Log in Harbor's UI as an admin user.
2. Click on the admin's name and select **About** from drop-down menu. 
3. Click on the **Download** link to save the certificate file as `ca.crt`.
4. Transmit the certificate file to a Docker host, put it under the below directory, you may need to create the directory if it does not exist:
   ```
      /etc/docker/certs.d/<host_name_or_IP_of_Harbor>/ca.crt
   ```
5. Restart Docker service.
6. Run `docker login` to verify that HTTPS is working.

To import the CA's certificate into VCH, complete Step 1-3 and refer to VCH's document for instructions.

## Reconfiguration
If you want to change the properties of Harbor, follow the below steps:  

1. **Power off** Harbor's virtual appliance.  
2. Right click on the VM and select "Edit Settings".  

 ![ova](img/ova/edit_settings.png)

3. Click the "vApp Options" tab, update the properties and  click "OK".  

 ![ova](img/ova/vapp_options.png)

4. **Power on** the VM.  

**Notes:**  
1. The authentication mode can only be set once on firtst boot. Subsequent modification of this option does not have any effect.  
2. The initial admin password, root password of the virtual appliance, MySQL root password, and all networking properties can not be modified using this method after Harbor's first launch. Modify them by the following approach:
 * **Harbor Admin Password**: Change it in Harbor admin portal.  
 * **Root Password of Virtual Appliance**: Change it by logging in the virtual appliance and doing it in the Linux operating system.  
 * **MySQL Root Password**: Change it by logging in the virtual appliance and doing it in the Linux operating system.  
 * **Networking Properties**: Visit `https://harbor_ip_address:5480`, log in with root/password of your virtual appliance and modify networking properties. Reboot the system after you changing them.  