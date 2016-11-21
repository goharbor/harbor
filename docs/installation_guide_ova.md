# Installing and Configuring Harbor on vSphere as Virtual Appliance

This guide walks you through the steps about installing and configuring Harbor on vSphere as an virtual appliance (OVA). If you are installing Harbor on a Linux host, refer to this **[Installation Guide](installation_guide.md)**.

## Prerequisites
* vCenter 5.x+ and at least an ESX host. 
* 2 vCPUs, 4GB memory and 100GB free disk space in datastore.
* A network with DHCP capability, or a static IP address for the virtual appliance.

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

9. Customize the properties of Harbor. The properties are described below. Note that at the very least, you just need to set the **Root Password**, **Harbor Admin Password** and **Database Password** properties.  

 ![ova](img/ova/ova08.png)

 * System
	* **Root Password**: The initial password of the root user. Subsequent changes of password should be performed in operating system. (8-128 characters)
	* **Harbor Admin Password**: The initial password of Harbor admin. It only works for the first time when Harbor starts. It has no effect after the first launch of Harbor. Change the admin password from UI after launching Harbor. 
	* **Database Password**: The initial password of the root user of MySQL database. Subsequent changes of password should be performed in operating system. (8-128 characters)
	* **Permit Root Login**: Specifies whether root use can log in using SSH.
	* **Self Registration**: Determine whether the self-registration is allowed or not. Set this to off to disable a user's self-registration in Harbor. This flag has no effect when users are stored in LDAP or AD.
	* **Garbage Collection**: When setting this to true, Harbor performs garbage collection everytime it boots up. The first time setting this flag to true needs to power off the VM and power it on again.

 * Authentication
	* **Authentication Mode**: The default authentication mode is db_auth. Set it to ldap_auth when users' credentials are stored in an LDAP or AD server. Note: this option can only be set once.
	* **LDAP URL**: The URL of an LDAP/AD server.
	* **LDAP Search DN**: A user's DN who has the permission to search the LDAP/AD server. If your LDAP/AD server does not support anonymous search, you should configure this DN and LDAP Seach Password.
	* **LDAP Search Password**: The password of the user for LDAP search.
	* **LDAP Base DN**: The base DN from which to look up a user in LDAP/AD.
	* **LDAP UID**: The attribute used in a search to match a user, it could be uid, cn, email, sAMAccountName or other attributes depending on your LDAP/AD server.

 * Security
	* **Protocol**: The protocol for accessing Harbor. Warning: setting it to http makes the communication insecure.
	* **SSL Cert**: Paste in the content of a certificate file. Leave blank for a generated self-signed certificate.
	* **SSL Cert Key**: Paste in the content of certificate key file. Leave blank for a generated key.
	* **Verify Remote Cert**: Determine whether the image replication should verify the certificate when it connects to a remote registry via TLS. Set this flag to off when the remote registry uses a self-signed or untrusted certificate.

 * Email Settings
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
	* **Network 1 IP Adress**: The IP address of this interface. Leave blank if DHCP is desired.
	* **Network 1 Netmask**: The netmask or prefix for this interface. Leave blank if DHCP is desired.

 **Notes:** If you want to enable HTTPS with a self-signed certificate created manually, refer to the "Getting a certificate" part of this [guide](https://github.com/vmware/harbor/blob/master/docs/configure_https.md#getting-a-certificate) for generating a certificate.  

 After you complete the properties, click "Next".  

10. Review your settings and click "Finish" to complete the deployment.  

 ![ova](img/ova/ova09.png)

11. Power on the virtual appliance. It may take a few minutes for the first bootup. The virtual appliance needs to initialize itself for configuration like netowrk address and password. 

12. When the appliance is ready, check from vSphere Web Client for its IP address. Open a browser and type in the URL `http(s)://harbor_ip_address` or `http(s)://harbor_host_name`. Log in as the admin user and verify Harbor has been successfully installed. 

13. For information on how to use Harbor, please refer to [User Guide of Harbor](user_guide.md).

## Reconfiguration
If you want to change the properties of Harbor, follow the below steps:  

1. **Power off** Harbor's virtual appliance.  
2. Right click on the VM and select "Edit Settings".  

 ![ova](img/ova/edit_settings.png)

3. Click the "vApp Options" tab, update the properties and  click "OK".  

 ![ova](img/ova/vapp_options.png)

4. **Power on** the VM.  

**Notes:**  
1. The authentication mode can only be set once on firtst boot. So subsequent modification of this option will have no effect.  
2. The initial admin password, root password of the virtual appliance, MySQL root password, and all networking properties can not be modified using this method after Harbor's first launch. Modify them by the following steps:
 * Harbor Admin Password: Change it in Harbor admin portal.  
 * Root Password of Virtual Appliance: Change it by logging in the virtual appliance and doing it in the Linux operating system.  
 * MySQL Root Password: Change it by logging in the virtual appliance and doing it in the Linux operating system.  
 * Networking Properties: Visit `https://harbor_ip_address:5480`, login with root/password of your virtual appliance and modify networking properties. Reboot the system after you changing them.  