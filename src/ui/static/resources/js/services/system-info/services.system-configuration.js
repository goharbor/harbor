/*
    Copyright (c) 2016 VMware, Inc. All Rights Reserved.
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
        
        http://www.apache.org/licenses/LICENSE-2.0
        
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/
(function() {
  
  'use strict';
  angular.module('harbor.services.system.info')
    .value('mockConf', mockConf)
    .service('ConfigurationService', ConfigurationService);
    
  function mockConf() {
    return {
      "auth_mode": {
        "value": "ldap_auth",
        "editable": true
      },
      "email_from": {
        "value": "admin \u003csample_admin@mydomain.com\u003e",
        "editable": true
      },
      "email_server": {
        "value": "smtp.mydomain.com",
        "editable": true
      },
      "email_server_port": {
        "value": 25,
        "editable": true
      },
      "email_ssl": {
        "value": "true",
        "editable": true
      },
      "email_username": {
        "value": "sample_admin@mydomain.com",
        "editable": true
      },
      "ldap_base_dn": {
        "value": "dc=mydomain,dc=com",
        "editable": true
      },
      "ldap_search_dn": {
        "value": "uid=tester,ou=people,dc=mydomain,dc=com",
        "editable": true
      },
      "ldap_uid": {
        "value": "cn",
        "editable": true
      },
      "ldap_filter": {
        "value": "(&(objectClass=*))",
        "editable": true
      },
      "ldap_url": {
        "value": "ldap.mydomain.com",
        "editable": true
      },
      "ldap_connection_timeout": {
        "value": 10,
        "editable": true
      },
      "ldap_scope": {
        "value": 2,
        "editable": true
      },
      "max_job_workers": {
        "value": 3,
        "editable": true
      },
      "project_creation_restriction": {
        "value": "adminonly",
        "editable": true
      },
      "self_registration": {
        "value": "off",
        "editable": true
      },
      "verify_remote_cert": {
        "value": "on",
        "editable": true
      }
    };
  }
 
  ConfigurationService.$inject = ['$http', '$q', '$timeout', 'mockConf'];
    
  function ConfigurationService($http, $q, $timeout, mockConf) {
    this.configuration = configuration;
    this.pingLDAP = pingLDAP;
    
    function configuration() {
      var deferred = $q.defer();
      $timeout(function() {
        if(true) {
          deferred.resolve(mockConf());
        }else{
          deferred.reject('error');
        }
      }, 50);
      return deferred.promise;
    }
    
    function pingLDAP(ldapConf) {
      return $http
        .post('/api/ldap/ping', ldapConf);
    }
    
  }
  
})();