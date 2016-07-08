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
    .module('harbor.app') 
    .factory('currentUser', currentUser)
    .factory('currentProjectMember', currentProjectMember);
  
  currentUser.$inject = ['$cookies', '$timeout'];
  
  function currentUser($cookies, $timeout) {
    return {
      set: function(user) {
        $cookies.putObject('user', user, {'path': '/'});
      },
      get: function() {
        return $cookies.getObject('user');
      },
      unset: function() {
        $cookies.remove('user', {'path': '/'});
      }
    };
  }  
  
  currentProjectMember.$inject = ['$cookies'];
  
  function currentProjectMember($cookies) {
    return {
      set: function(member) {
        $cookies.putObject('member', member, {'path': '/'});
      },
      get: function() {
        return $cookies.getObject('member');
      },
      unset: function() {
        $cookies.remove('member', {'path': '/'});
      }
    };
  }
      
})();