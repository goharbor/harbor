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
    .module('harbor.layout.change.password')
    .controller('ChangePasswordController', ChangePasswordController);
  
  ChangePasswordController.$inject = ['ChangePasswordService', 'UpdateUserService', '$filter', 'trFilter', '$scope', '$window', 'currentUser', 'getParameterByName', '$location'];
  
  function ChangePasswordController(ChangePasswordService, UpdateUserService, $filter, trFilter, $scope, $window, currentUser, getParameterByName, $location) {

    var vm = this;
    vm.isOpen = false;
 
    vm.hasError = false;
    vm.errorMessage = '';
    
    vm.reset = reset;    
    
    vm.confirm = confirm;
    vm.updatePassword = updatePassword;
    vm.cancel = cancel;
        
    var userId = getParameterByName('user_id', $location.absUrl());
    var username = getParameterByName('username', $location.absUrl());
    var returnUrl = '/admin_option';
   
    vm.titleChangePassword = $filter('tr')('change_user_password', [username]);
        
    if (!userId) {
      $scope.user = currentUser.get();
      if(!$scope.user) {
        $window.location.href = '/';
        return;
      }
      userId = $scope.user.user_id;
            
      vm.titleChangePassword = $filter('tr')('change_password', []);
      returnUrl = '/dashboard';
    }
    
   
    
    //Error message dialog handler for account setting.
    $scope.$on('modalTitle', function(e, val) {
      vm.modalTitle = val;
    });
    
    $scope.$on('modalMessage', function(e, val) {
      vm.modalMessage = val;
    });
       
    $scope.$on('raiseError', function(e, val) {
      if(val) {   
        vm.action = function() {
          $scope.$broadcast('showDialog', false);
        };
        vm.contentType = 'text/plain';    
        vm.confirmOnly = true;  
        $scope.$broadcast('showDialog', true);
      }
    });
    
    
    function getUserSuccess(response) {
      $scope.user = response.data;
    }
    
    function getUserFailed(response) {
      vm.modalTitle = $filter('tr')('change_profile', []);
      vm.modalMessage = $filter('tr')('failed_to_update_user', []);
      $scope.$broadcast('showDialog', true);
      console.log('Failed to get user.');
    }
        
    function reset() {
      $scope.form.$setUntouched();
      $scope.form.$setPristine();
      vm.hasError = false;
      vm.errorMessage = '';
    }
     
    function confirm() {     
      $window.location.href = returnUrl;
    }    
                
    function updatePassword(user) {
      if(user && angular.isDefined(user.oldPassword) && angular.isDefined(user.password)) {
        vm.action = vm.confirm;
        ChangePasswordService(userId, user.oldPassword, user.password)
          .success(changePasswordSuccess)
          .error(changePasswordFailed);
      }
     
    }
    
    function changePasswordSuccess(data, status) {
      vm.modalTitle = $filter('tr')('change_password', []);
      vm.modalMessage = $filter('tr')('successful_changed_password', []);
      $scope.$broadcast('showDialog', true);
    }
    
    function changePasswordFailed(data, status) {

      var message;
      $scope.$emit('modalTitle', $filter('tr')('error'));
      console.log('Failed to change password:' + data);
      if(data === 'old_password_is_not_correct') {
        message = $filter('tr')('old_password_is_incorrect');
      }else{
        message = $filter('tr')('failed_to_change_password');
      }

      $scope.$emit('modalMessage', message);
      $scope.$emit('raiseError', true);
    }
    
    function cancel(form) {
      $window.location.href = returnUrl;
    }
    
  }
  
})();
