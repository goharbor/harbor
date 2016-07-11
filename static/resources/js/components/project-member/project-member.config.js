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
    .module('harbor.project.member')
    .constant('roles', roles)
    .factory('getRole', getRole);
    
  function roles() {
    return [
      {'id': '1', 'name': 'project_admin', 'roleName': 'projectAdmin'},
      {'id': '2', 'name': 'developer', 'roleName': 'developer'},
      {'id': '3', 'name': 'guest', 'roleName': 'guest'}
    ];
  }
  
  getRole.$inject = ['roles', '$filter', 'trFilter'];
  
  function getRole(roles, $filter, trFilter) {
    var r = roles();
    return get;     
    function get(query) {
     
      for(var i = 0; i < r.length; i++) {
        var role = r[i];
        if(query.key === 'roleName' && role.roleName === query.value || query.key === 'roleId' && role.id === String(query.value)) {
           return role;
        }
      }
    }
  }
})();
