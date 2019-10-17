# Configure Authentication

You can change authentication mode between **Database**(default) and **LDAP** before any user is added, when there is at least one user(besides admin) in Harbor, you cannot change the authentication mode.  
![browse project](../img/new_auth.png)
When using LDAP mode, user's self-registration is disabled. The parameters of LDAP server must be filled in. For more information, refer to [User account](#user-account).   
![browse project](../img/ldap_auth.png)

When using OIDC mode, user will login Harbor via OIDC based SSO.  A client has to be registered on the OIDC provider and Harbor's callback URI needs to be associated to that client as a redirectURI.
![OIDC settings](../img/oidc_auth_setting.png)

The settings of this auth mode:
* OIDC Provider Name: The name of the OIDC Provider.
* OIDC Provider Endpoint: The URL of the endpoint of the OIDC provider(a.k.a the Authorization Server in OAuth's terminology), 
which must service the "well-known" URI for its configuration, more details please refer to https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationRequest
* OIDC Client ID: The ID of client configured on OIDC Provider.
* OIDC Client Secret: The secret for this client.
* OIDC Scope: The scope values to be used during the authentication.  It is the comma separated string, which must contain `openid`.  
Normally it should also contain `profile` and `email`.  For getting the refresh token it should also contain `offline_access`.  Please check with the administrator of the OIDC Provider.
* Verify Certificate: Whether to check the certificate when accessing the OIDC Provider. if you are running the OIDC Provider with self-signed
certificate, make sure this value is set to false.

## User account
Harbor supports different authentication modes:  

* **Database(db_auth)**  

	Users are stored in the local database.  
	
	A user can register himself/herself in Harbor in this mode. To disable user self-registration, refer to the [installation guide](installation_guide.md) for initial configuration, or disable this feature in [Administrator Options](#administrator-options). When self-registration is disabled, the system administrator can add users into Harbor.  
	
	When registering or adding a new user, the username and email must be unique in the Harbor system. The password must contain at least 8 characters with 1 lowercase letter, 1 uppercase letter and 1 numeric character.
	
	When you forgot your password, you can follow the below steps to reset the password:  

	1. Click the link "Forgot Password" in the sign in page.  
	2. Input the email address entered when you signed up, an email will be sent out to you for password reset.  
	3. After receiving the email, click on the link in the email which directs you to a password reset web page.  
	4. Input your new password and click "Save".  
	
* **LDAP/Active Directory (ldap_auth)**  

	Under this authentication mode, users whose credentials are stored in an external LDAP or AD server can log in to Harbor directly.  
	
	When an LDAP/AD user logs in by *username* and *password*, Harbor binds to the LDAP/AD server with the **"LDAP Search DN"** and **"LDAP Search Password"** described in [installation guide](installation_guide.md). If it succeeded, Harbor looks up the user under the LDAP entry **"LDAP Base DN"** including substree. The attribute (such as uid, cn) specified by **"LDAP UID"** is used to match a user with the *username*. If a match is found, the user's *password* is verified by a bind request to the LDAP/AD server. Uncheck **"LDAP Verify Cert"** if the LDAP/AD server uses a self-signed or an untrusted certificate.
	
	Self-registration, deleting user, changing password and resetting password are not supported under LDAP/AD authentication mode because the users are managed by LDAP or AD.  

* **OIDC Provider (oidc_auth)**

    With this authentication mode, regular user will login to Harbor Portal via SSO flow.  
    After the system administrator configure Harbor to authenticate via OIDC (more details refer to [this section](#managing-authentication)),
    a button `LOGIN VIA OIDC PROVIDER` will appear on the login page.  
    ![oidc_login](../img/oidc_login.png)
    
    By clicking this button user will kick off the SSO flow and be redirected to the OIDC Provider for authentication.  After a successful
    authentication at the remote site, user will be redirected to Harbor.  There will be an "onboard" step if it's the first time the user 
    authenticate using his account, in which there will be a dialog popped up for him to set his user name in Harbor:
    ![oidc_onboar](../img/oidc_onboard_dlg.png)
    
    This user name will be the identifier for this user in Harbor, which will be used in the cases such as adding member to a project, assigning roles, etc.
    This has to be a unique user name, if another user has used this user name to onboard, user will be prompted to choose another one.
    
    Regarding this user to use docker CLI, please refer to [Using CLI after login via OIDC based SSO](#using-oidc-cli-secret)
   
    **NOTE:**
    1. After the onboard process, you still have to login to Harbor via SSO flow, the `Username` and `Password` fields are only for
    local admin to login when Harbor is configured authentication via OIDC.
    2. Similar to LDAP authentication mode, self-registration, updating profile, deleting user, changing password and 
    resetting password are not supported.
    
## Using OIDC CLI secret

Having authenticated via OIDC SSO and onboarded to Harbor, you can use Docker/Helm CLI to access Harbor to read/write the artifacts.
As the CLI cannot handle redirection for SSO, we introduced `CLI secret`, which is only available when Harbor's authentication mode 
is configured to OIDC based.  
After logging into Harbor, click the drop down list to view user's profile:
![user_profile](../img/user_profile.png)

You can copy your CLI secret via the dialog of profile:
![profile_dlg](../img/profile_dlg.png)

After that you can authenticate using your user name in Harbor that you set during onboard process, and CLI secret as the password
with Docker/Helm CLI, for example:
```sh
docker login -u testuser -p xxxxxx jt-test.local.goharbor.io

``` 

When you click the "..." icon in the profile dialog, a button for generating new CLI secret will appear, and you can generate a new 
CLI secret by clicking this button.  Please be reminded one user can only have one CLI secret, so when a new secret is generated, the
old one becomes invalid at once.

**NOTE**:
Under the hood the CLI secret is associated with the ID token, and Harbor will try to refresh the token, so the CLI secret will
be valid after th ID token expires. However, if the OIDC Provider does not provide refresh token or the refresh fails for some 
reason, the CLI secret will become invalid.  In that case you can logout and login Harbor via SSO flow again so Harbor can get a 
new ID token and the CLI secret will work again.
Title

Text
