[Back to table of contents](../../_index.md)

----------

# Authentication Modes and User Accounts
Harbor supports different modes for authenticating users and managing user accounts.

- [Database Authentication](#db_auth)
- [LDAP/Active Directory Authentication](#ldap_auth)
- [OIDC Provider Authentication](#oidc_auth)

**NOTE**: The Harbor interface offers an option to configure UAA authentication. This authentication mode is not recommended and is not documented in this guide.

<a id="db_auth"></a>
## Database Authentication

In database authentication mode, user accounts are stored in the local database. By default, only the Harbor system administrator can create user accounts to add users to Harbor. You can optionally configure Harbor to allow self-registration.  

**IMPORTANT**: If you create users in the database, Harbor is locked in database mode. You cannot change to a different authentication mode after you have created local users.

1. Log in to the Harbor interface with an account that has Harbor system administrator privileges.
1. Under **Administration**, go to **Configuration** and select the **Authentication** tab.
1. Leave **Auth Mode** set to the default **Database** option.

   ![Database authentication](../../img/db_auth.png)
   
1. Optionally select the **Allow Self-Registration** check box.

   ![Enable self-registration](../../img/new_self_reg.png)
    
   If you enable self registration option, users can register themselves in Harbor. Self-registration is disabled by default. If you enable self-registration, unregistered users can sign up for a Harbor account by clicking **Sign up for an account** in the Harbor log in page.
    
    ![Enable self-registration](../../img/self-registration-login.png)

<a id="ldap_auth"></a>
## LDAP/Active Directory Authentication 

If you select LDAP/AD authentication, users whose credentials are stored in an external LDAP or AD server can log in to Harbor directly. In this case, you do not create user accounts in Harbor.

**IMPORTANT**: You can change the authentication mode from database to LDAP only if no local users have been added to the database. If there is at least one user other than `admin` in the Harbor database, you cannot change the authentication mode.

Because the users are managed by LDAP or AD, self-registration, creating users, deleting users, changing passwords, and resetting passwords are not supported in LDAP/AD authentication mode.  

If you want to manage user authentication by using LDAP groups, you must enable the `memberof` feature on the LDAP/AD server. With the `memberof` feature, the LDAP/AD user entity's `memberof` attribute is updated when the group entity's `member` attribute is updated, for example by adding or removing an LDAP/AD user from the LDAP/AD group. This feature is enabled by default in Active Directory. For information about how to enable and verify `memberof` overlay in OpenLDAP, see [this technical note]( https://technicalnotes.wordpress.com/2014/04/19/openldap-setup-with-memberof-overlay/).

1. Log in to the Harbor interface with an account that has Harbor system administrator privileges.
1. Under **Administration**, go to **Configuration** and select the **Authentication** tab.
1. Use the **Auth Mode** drop-down menu to select **LDAP**.

   ![LDAP authentication](../../img/select_ldap_auth.png)
1. Enter the address of your LDAP server, for example `ldaps://10.162.16.194`.
1. Enter information about your LDAP server.

   - **LDAP Search DN** and **LDAP Search Password**: When a user logs in to Harbor with their LDAP username and password, Harbor uses these values to bind to the LDAP/AD server. For example, `cn=admin,dc=example.com`.
   - **LDAP Base DN**: Harbor looks up the user under the LDAP Base DN entry, including the subtree. For example, `dc=example.com`.
   - **LDAP Filter**: The filter to search for LDAP/AD users. For example, `objectclass=user`. 
   - **LDAP UID**: An attribute, for example `uid`, or `cn`, that is used to match a user with the username. If a match is found, the user's password is verified by a bind request to the LDAP/AD server. 
   - **LDAP Scope**: The scope to search for LDAP/AD users. Select from **Subtree**, **Base**, and **OneLevel**.
   
     ![Basic LDAP configuration](../../img/ldap_auth.png)  
1. If you want to manage user authentication with LDAP groups, configure the group settings.
   - **LDAP Group Base DN**: The base DN from which to lookup a group in LDAP/AD. For example, `ou=groups,dc=example,dc=com`.
   - **LDAP Group Filter**: The filter to search for LDAP/AD groups. For example, `objectclass=groupOfNames`. 
   - **LDAP Group GID**: The attribute used to name an LDAP/AD group. For example, `cn`.  
   - **LDAP Group Admin DN**: All LDAP/AD users in this group DN have Harbor system administrator privileges.
   - **LDAP Group Membership**: The user attribute usd to identify a user as a member of a group. By default this is `memberof`.
   - **LDAP Scope**: The scope to search for LDAP/AD groups. Select from **Subtree**, **Base**, and **OneLevel**.
   
     ![LDAP group configuration](../../img/ldap_groups.png)
1. Uncheck **LDAP Verify Cert** if the LDAP/AD server uses a self-signed or untrusted certificate.

   ![LDAP certificate verification](../../img/ldap_cert_test.png)
1. Click **Test LDAP Server** to make sure that your configuration is correct.
1. Click **Save** to complete the configuration.

<a id="oidc_auth"></a>
## OIDC Provider Authentication

If you select OpenID Connect (OIDC) authentication, users log in to the Harbor interface via an OIDC single sign-on (SSO) provider, such as Okta, KeyCloak, or dex. In this case, you do not create user accounts in Harbor.

**IMPORTANT**: You can change the authentication mode from database to OIDC only if no local users have been added to the database. If there is at least one user other than `admin` in the Harbor database, you cannot change the authentication mode.

Because the users are managed by the OIDC provider, self-registration, creating users, deleting users, changing passwords, and resetting passwords are not supported in OIDC authentication mode.  

### Configure Your OIDC Provider

You must configure your OIDC provider so that you can use it with Harbor. For precise information about how to perform these configurations, see the documentation for your OIDC provider.

- Set up the users and groups that will use the OIDC provider to log in to Harbor. You do not need to assign any specific OIDC roles to users or groups as these do not get mapped to Harbor roles.
- The URL of the OIDC provider endpoint, known as the Authorization Server in OAuth terminology, must service the well-known URI for its configuration document. For more information about the configuration document, see the [OpenID documentation] (https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationRequest).
- To manage users by using OIDC groups, create a custom group claim that contains all of the user groups that you want to register in Harbor. The group claim must be mapped in the ID token that is sent to Harbor when users log in. You can enable the `memberof` feature on the OIDC provider. With the `memberof` feature, the OIDC user entity's `memberof` attribute is updated when the group entity's `member` attribute is updated, for example by adding or removing an OIDC user from the OIDC group.
- Register Harbor as a client application with the OIDC provider. Associate Harbor's callback URI to the client application as a `redirectURI`. This is the address to which the OIDC provider sends ID tokens.

### Configure an OIDC Provider in Harbor

Before configuring an OIDC provider in Harbor, make sure that your provider is configured correctly according to the preceding section.

1. Log in to the Harbor interface with an account that has Harbor system administrator privileges.
1. Under **Administration**, go to **Configuration** and select the **Authentication** tab.
1. Use the **Auth Mode** drop-down menu to select **OIDC**.

   ![LDAP authentication](../../img/select_oidc_auth.png)
1. Enter information about your OIDC provider.   

   - **OIDC Provider Name**: The name of the OIDC provider.
   - **OIDC Provider Endpoint**: The URL of the endpoint of the OIDC provider.
   - **OIDC Client ID**: The client ID with which Harbor is registered as  client application with the OIDC provider.
   - **OIDC Client Secret**: The secret for the Harbor client application.
   - **Group Claim Name**: The name of a custom group claim that you have configured in your OIDC provider, that includes the groups to add to Harbor.
   - **OIDC Scope**: A comma-separated string listing the scopes to be used during authentication. 
   
       The OIDC scope must contain `openid` and usually also contains `profile` and `email`. To obtain refresh tokens it should also contain `offline_access`. If you are using OIDC groups, a scope must identify the group claim. Check with your OIDC provider administrator for precise details of how to identify the group claim scope, as this differs from vendor to vendor.
       
       ![OIDC settings](../../img/oidc_auth_setting.png)
1. Uncheck **Verify Certificate** if the OIDC Provider uses a self-signed or untrusted certificate.
1. Verify that the Redirect URI that you configured in your OIDC provider is the same as the one displayed at the bottom of the page. 
      
     ![OIDC certificate verification, URI, and test ](../../img/oidc_cert_verification.png)
1. Click **Test OIDC Server** to make sure that your configuration is correct.
1. Click **Save** to complete the configuration.

### Log In to Harbor via an OIDC Provider

When the Harbor system administrator has configured Harbor to authenticate via OIDC a **Login via OIDC Provider** button appears on the Harbor login page.  

![oidc_login](../../img/oidc_login.png)

**NOTE:** When Harbor is configured authentication via OIDC, the **Username** and **Password** fields are reserved for the local Harbor system administrator to log in.
    
1. As a Harbor user, click the **Login via OIDC Provider** button.
 
   This redirects you to the OIDC Provider for authentication.  
1. If this is the first time that you are logging in to Harbor with OIDC, specify a user name for Harbor to associate with your OIDC username.

   ![Specify Harbor username for OIDC](../../img/oidc_onboard_dlg.png)
    
   This is the user name by which you are identified in Harbor, which is used when adding you to projects, assigning roles, and so on. If the username is already taken, you are prompted to choose another one.
1. After the OIDC provider has authenticated you, you are redirected back to Harbor.

### Using OIDC from the Docker or Helm CLI

After you have authenticated via OIDC and logged into the Harbor interface for the first time, you can use the Docker or Helm CLI to access Harbor.

The Docker and Helm CLIs cannot handle redirection for OIDC, so Harbor provides a CLI secret for use when logging in from Docker or Helm. This is only available when Harbor uses OIDC authentication.  

1. Log in to Harbor with an OIDC user account.
1. Click your username at the top of the screen and select **User Profile**.

   ![Access user profile](../../img/user_profile.png)
1. Click the clipboard icon to copy the CLI secret associated with your account.

   ![Copy CLI secret](../../img/profile_dlg.png)
1. Optionally click the **...** icon in your user profile to display buttons for automatically generating or manually creating a new CLI secret.

   ![Copy CLI secret](../../img/generate_create_new_secret.png) 

   A user can only have one CLI secret, so when a new secret is generated or create, the old one becomes invalid.
1. If you generated a new CLI secret, click the clipboard icon to copy it.

You can now use your CLI secret as the password when logging in to Harbor from the Docker or Helm CLI.

<pre>
sh docker login -u testuser -p <i>cli_secret</i> jt-test.local.goharbor.io
</pre> 

**NOTE**: The CLI secret is associated with the OIDC ID token. Harbor will try to refresh the token, so the CLI secret will be valid after the ID token expires. However, if the OIDC Provider does not provide a refresh token or the refresh fails, the CLI secret becomes invalid. In this case, log out and log back in to Harbor via your OIDC provider so that Harbor can get a new ID token. The CLI secret will then work again.

----------

[Back to table of contents](../../_index.md)