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
    .module('harbor.session')
    .controller('CurrentUserController', CurrentUserController);
 
  CurrentUserController.$inject = ['$scope', 'CurrentUserService', 'currentUser', '$window', '$document'];
  
  function CurrentUserController($scope, CurrentUserService, currentUser, $window, $document) {
    
    var vm = this;
         
    CurrentUserService()
      .then(getCurrentUserComplete)
      .catch(getCurrentUserFailed);
        
    function getCurrentUserComplete(response) {
      if(angular.isDefined(response)) {
        currentUser.set(response.data);  
        if(location.pathname === '/') {
          $window.location.href = '/dashboard';
        }
      }   
    }
    
    function getCurrentUserFailed(e){
      console.log('No session of current user.');
    }   
  }
 
})();