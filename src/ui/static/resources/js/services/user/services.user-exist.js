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
    .module('harbor.services.user')
    .factory('UserExistService', UserExistService);
    
  UserExistService.$inject = ['$http', '$log'];
   
  function UserExistService($http, $log) {
    return userExist;
    function userExist(target, value) {
      return  $.ajax({
          type: 'POST',
          url: '/userExists',
          async: false,
          data: {'target': target, 'value': value}
      });
    } 
  }
  
})();