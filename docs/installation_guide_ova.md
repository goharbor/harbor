# Install and Configure Harbor on vSphere using OVA
This guide takes you through the steps about installing and configuring Harbor on vSphere using OVA.

## Installation
1.Get URL or download the OVA file to your local disk from [release page](https://github.com/vmware/harbor/releases).  

2.Login vSphere web client. Right click on the datacenter, cluster or host which Harbor will be deployed on. Select "Deploy OVF Template" and open the import wizard.  

![ova](img/ova/ova01.png)

3.Paste the URL of OVA file or select it from local disk and click "Next".  

![ova](img/ova/ova02.png)

4.Review the OVF template details and click "Next".  

![ova](img/ova/ova03.png)

5.Spefify a name and location for the deployed template.  

![ova](img/ova/ova04.png)

6.Select the storage and virtual disk format, click "Next".  

![ova](img/ova/ova05.png)

7.Configure the networks the deployed template should use.  

![ova](img/ova/ova06.png)

8.Customize the properties of Harbor. The properties are described below. Note that at the very least, you just need to set the **Root Password**, **Harbor Admin Password** and **Database Password** properties.  

![ova](img/ova/ova07.png)

* Application
	* **Root Password**: The password of the root user. (8-128 characters)
	* **Harbor Admin Password**: The initial password of Harbor admin. It only works for the first time when Harbor starts. It has no effect after the first launch of Harbor. Change the admin password from UI after launching Harbor. (8-20 characters)
	* **Database Password**: The password of the root user of MySQL database.  (8-128 characters)
	* **Authentication Mode**: The default authentication mode is db_auth, i.e. the credentials are stored in a local database. Set it to ldap_auth if you want to verify the user's credential against an LDAP/AD server.
	* **LDAP URL**: The URL of an LDAP/AD server.
	* **LDAP Search DN**: A user's DN who has the permission to search the LDAP/AD server. If your LDAP/AD server does not support anonymous search, you should configure this DN and LDAP Seach Password.
	* **LDAP Search Password**: The password of the user for LDAP search.
	* **LDAP Base DN**: The base DN from which to look up a user in LDAP/AD.
	* **LDAP UID**: The attribute used in a search to match a user, it could be uid, cn, email, sAMAccountName or other attributes depending on your LDAP/AD server.
	* **Email Server**: The mail server to send out emails to reset password. 
	* **Email Server Port**: The port of mail server.
	* **Email Username**: The user from whom the password reset email is sent.
	* **Email Password**: The password of the user from whom the password reset email is sent.
	* **Email From**: The name of the email sender.
	* **Email SSL**: Whether to enabled secure mail transmission.
	* **SSL Cert**: Paste in the content of a certificate file. If SSL Cert and SSL Cert Key are both set, HTTPS will be used.
	* **SSL Cert Key**: Paste in the content of certificate key file. If SSL Cert and SSL Cert Key are both set, HTTPS will be used.
	* **Verify Remote Cert**: Determine whether the image replication should verify the SSL certificate when it connects to a remote registry. Set this flag to off when the remote registry uses a self-signed or untrusted certificate.
	* **Garbage Collection**: When setting this to true, Harbor performs garbage collection everytime it boots up.

* Networking properties
	* **Default Gateway**: The default gateway address for this VM. Leave blank if DHCP is desired.
	* **Domain Name**: The domain name of this VM. Leave blank if DHCP is desired.
	* **Domain Search Path**: The domain search path(comma or space separated domain names) for this VM. Leave blank if DHCP is desired.
	* **Domain Name Servers**: The domain name server IP Address for this VM(comma separated). Leave blank if DHCP is desired.
	* **Network 1 IP Adress**: The IP address of this interface. Leave blank if DHCP is desired.
	* **Network 1 Netmask**: The netmask or prefix for this interface. Leave blank if DHCP is desired.

**Notes:** If you want to enable HTTPS with a self-signed certificate and have no idea how to generate it, refer to the "Getting a certificate" part of this [guide](https://github.com/vmware/harbor/blob/master/docs/configure_https.md#getting-a-certificate).  

After you complete the properties, click "Next".  

9.Review your settings and click "Finish" to complete the installation.  

![ova](img/ova/ova08.png)

## Reconfiguration
If you want to reconfigure the properties of Harbor, follow the steps:  
1.Power off the VM which Harbor is deployed on.  
2.Right click on the VM and select "Edit Settings".  

![ova](img/ova/edit_settings.png)

3.Click the "vApp Options" tab, reconfigure the properties and  click "OK".  

![ova](img/ova/vapp_options.png)

4.Power on the VM.  

**Notes:** "Harbor Admin Password" and all networking properties can not be modified using this method after Harbor launched. Change the admin password from UI and change the networking properties in the OS level manually.  