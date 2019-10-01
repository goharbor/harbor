## Configuration file of Harbor

#This attribute is for migrator to detect the version of the .cfg file, DO NOT MODIFY!
_version = 1.7.0
#The IP address or hostname to access admin UI and registry service.
#DO NOT use localhost or 127.0.0.1, because Harbor needs to be accessed by external clients.
#DO NOT comment out this line, modify the value of "hostname" directly, or the installation will fail.
hostname = $hostname

#The protocol for accessing the UI and token/notification service, by default it is http.
#It can be set to https if ssl is enabled on nginx.
ui_url_protocol = $ui_url_protocol

#Maximum number of job workers in job service
max_job_workers = 10

#Determine whether or not to generate certificate for the registry's token.
#If the value is on, the prepare script creates new root cert and private key
#for generating token to access the registry. If the value is off the default key/cert will be used.
#This flag also controls the creation of the notary signer's cert.
customize_crt = $customize_crt

#The path of cert and key files for nginx, they are applied only the protocol is set to https
ssl_cert = $ssl_cert
ssl_cert_key = $ssl_cert_key

#The path of secretkey storage
secretkey_path = $secretkey_path

#Admiral's url, comment this attribute, or set its value to NA when Harbor is standalone
admiral_url = $admiral_url

#Log files are rotated log_rotate_count times before being removed. If count is 0, old versions are removed rather than rotated.
log_rotate_count = $log_rotate_count
#Log files are rotated only if they grow bigger than log_rotate_size bytes. If size is followed by k, the size is assumed to be in kilobytes.
#If the M is used, the size is in megabytes, and if G is used, the size is in gigabytes. So size 100, size 100k, size 100M and size 100G
#are all valid.
log_rotate_size = $log_rotate_size

#Config http proxy for Clair, e.g. http://my.proxy.com:3128
#Clair doesn't need to connect to harbor internal components via http proxy.
http_proxy = $http_proxy
https_proxy = $https_proxy
no_proxy = $no_proxy

#NOTES: The properties between BEGIN INITIAL PROPERTIES and END INITIAL PROPERTIES
#only take effect in the first boot, the subsequent changes of these properties
#should be performed on web ui

#************************BEGIN INITIAL PROPERTIES************************

#Email account settings for sending out password resetting emails.

#Email server uses the given username and password to authenticate on TLS connections to host and act as identity.
#Identity left blank to act as username.
email_identity =

email_server = smtp.mydomain.com
email_server_port = 25
email_username = sample_admin@mydomain.com
email_password = abc
email_from = admin <sample_admin@mydomain.com>
email_ssl = false
email_insecure = false

##The initial password of Harbor admin, only works for the first time when Harbor starts.
#It has no effect after the first launch of Harbor.
#Change the admin password from UI after launching Harbor.
harbor_admin_password = Harbor12345

##By default the auth mode is db_auth, i.e. the credentials are stored in a local database.
#Set it to ldap_auth if you want to verify a user's credentials against an LDAP server.
auth_mode = db_auth

#The url for an ldap endpoint.
ldap_url = ldaps://ldap.mydomain.com

#A user's DN who has the permission to search the LDAP/AD server.
#If your LDAP/AD server does not support anonymous search, you should configure this DN and ldap_search_pwd.
#ldap_searchdn = uid=searchuser,ou=people,dc=mydomain,dc=com

#the password of the ldap_searchdn
#ldap_search_pwd = password

#The base DN from which to look up a user in LDAP/AD
ldap_basedn = ou=people,dc=mydomain,dc=com

#Search filter for LDAP/AD, make sure the syntax of the filter is correct.
#ldap_filter = (objectClass=person)

# The attribute used in a search to match a user, it could be uid, cn, email, sAMAccountName or other attributes depending on your LDAP/AD
ldap_uid = uid

#the scope to search for users, 0-LDAP_SCOPE_BASE, 1-LDAP_SCOPE_ONELEVEL, 2-LDAP_SCOPE_SUBTREE
ldap_scope = 2

#Timeout (in seconds)  when connecting to an LDAP Server. The default value (and most reasonable) is 5 seconds.
ldap_timeout = 5

#Verify certificate from LDAP server
ldap_verify_cert = true

#The base dn from which to lookup a group in LDAP/AD
ldap_group_basedn = ou=group,dc=mydomain,dc=com

#filter to search LDAP/AD group
ldap_group_filter = objectclass=group

#The attribute used to name a LDAP/AD group, it could be cn, name
ldap_group_gid = cn

#The scope to search for ldap groups. 0-LDAP_SCOPE_BASE, 1-LDAP_SCOPE_ONELEVEL, 2-LDAP_SCOPE_SUBTREE
ldap_group_scope = 2

#Turn on or off the self-registration feature
self_registration = on

#The expiration time (in minute) of token created by token service, default is 30 minutes
token_expiration = 30

#The flag to control what users have permission to create projects
#The default value "everyone" allows everyone to creates a project.
#Set to "adminonly" so that only admin user can create project.
project_creation_restriction = everyone

#************************END INITIAL PROPERTIES************************

#######Harbor DB configuration section#######

#The address of the Harbor database. Only need to change when using external db.
db_host = $db_host

#The password for the root user of Harbor DB. Change this before any production use.
db_password = $db_password

#The port of Harbor database host
db_port = $db_port

#The user name of Harbor database
db_user = $db_user

##### End of Harbor DB configuration#######

##########Redis server configuration.############

#Redis connection address
redis_host = redis

#Redis connection port
redis_port = 6379

#Redis connection password
redis_password =

#Redis connection db index
#db_index 1,2,3 is for registry, jobservice and chartmuseum.
#db_index 0 is for UI, it's unchangeable
redis_db_index = 1,2,3

########## End of Redis server configuration ############

##########Clair DB configuration############

#Clair DB host address. Only change it when using an exteral DB.
clair_db_host = $clair_db_host
#The password of the Clair's postgres database. Only effective when Harbor is deployed with Clair.
#Please update it before deployment. Subsequent update will cause Clair's API server and Harbor unable to access Clair's database.
clair_db_password = $clair_db_password
#Clair DB connect port
clair_db_port = $clair_db_port
#Clair DB username
clair_db_username = $clair_db_username
#Clair default database
clair_db = $clair_db

#The interval of clair updaters, the unit is hour, set to 0 to disable the updaters.
clair_updaters_interval = 12

##########End of Clair DB configuration############

#The following attributes only need to be set when auth mode is uaa_auth
uaa_endpoint = $uaa_endpoint
uaa_clientid = $uaa_clientid
uaa_clientsecret = $uaa_clientsecret
uaa_verify_cert = $uaa_verify_cert
uaa_ca_cert = $uaa_ca_cert


### Harbor Storage settings ###
#Please be aware that the following storage settings will be applied to both docker registry and helm chart repository.
#registry_storage_provider can be: filesystem, s3, gcs, azure, etc.
registry_storage_provider_name = $registry_storage_provider_name
#registry_storage_provider_config is a comma separated "key: value" pairs, e.g. "key1: value, key2: value2".
#To avoid duplicated configurations, both docker registry and chart repository follow the same storage configuration specifications of docker registry.
#Refer to https://docs.docker.com/registry/configuration/#storage for all available configuration.
registry_storage_provider_config = $registry_storage_provider_config
#registry_custom_ca_bundle is the path to the custom root ca certificate, which will be injected into the truststore
#of registry's and chart repository's containers.  This is usually needed when the user hosts a internal storage with self signed certificate.
registry_custom_ca_bundle =

#If reload_config=true, all settings which present in harbor.cfg take effect after prepare and restart harbor, it overwrites exsiting settings.
#reload_config=true
#Regular expression to match skipped environment variables
#skip_reload_env_pattern=(^EMAIL.*)|(^LDAP.*)
