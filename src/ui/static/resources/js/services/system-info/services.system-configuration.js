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
    .service('ConfigurationService', ConfigurationService);
    
  ConfigurationService.$inject = ['$http', '$q', '$timeout'];
    
  function ConfigurationService($http, $q, $timeout) {
    this.get = get;
    this.update = update;
    this.pingLDAP = pingLDAP;
    
    function get() {
      return $http.get('/api/configurations');
    }
    
    function update(updates) {
      return $http.put('/api/configurations', updates);
    }
    
    function pingLDAP(ldapConf) {
      return $http
        .post('/api/ldap/ping', ldapConf);
    }
    
  }
  
})();