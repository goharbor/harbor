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
  
  angular
    .module('harbor.services.destination')
    .factory('PingDestinationService', PingDestinationService);
    
  PingDestinationService.$inject = ['$http'];
  
  function PingDestinationService($http) {
    return pingDestination;
    function pingDestination(target) {
      var payload = {};
      if(target['id']) {
        payload = {'id': target['id']};
      }else {
        payload = {
          'name': target['name'],
          'endpoint': target['endpoint'],
          'username': target['username'],
          'password': target['password']
        };
      }
      
      return $http({
           'method': 'POST',
              'url': '/api/targets/ping',
          'headers': {'Content-Type': 'application/x-www-form-urlencoded'},
          'transformRequest': function(obj) {
              var str = [];
              for(var p in obj) {
                str.push(encodeURIComponent(p) + "=" + encodeURIComponent(obj[p]));
              }
              return str.join("&");
          },
          'timeout': 30000,
          'data': payload
        });
    }
  }
  
})();