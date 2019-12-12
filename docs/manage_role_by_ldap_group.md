## Introduction

You can import an LDAP/AD group to Harbor and assign project roles to it. All LDAP/AD users in this LDAP/AD group have assigned roles.

**NOTE**: Information about how to configure LDAP Groups in the Harbor interface has migrated to the [Harbor User Guide](user_guide.md).
    
To configure LDAP parameters via the API, see **[Configure Harbor User Settings from the Command Line](configure_user_settings.md)**

For example:
```
curl -X PUT -u "<username>:<password>" -H "Content-Type: application/json" -ki https://harbor.sample.domain/api/configurations -d'{"ldap_group_basedn":"ou=groups,dc=example,dc=com"}'
```
