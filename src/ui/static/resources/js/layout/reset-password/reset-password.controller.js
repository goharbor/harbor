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
    .module('harbor.layout.reset.password')
    .controller('ResetPasswordController', ResetPasswordController);
  
  ResetPasswordController.$inject = ['$scope', '$location', 'ResetPasswordService', '$window', 'getParameterByName', '$filter', 'trFilter'];
  
  function ResetPasswordController($scope, $location, ResetPasswordService, $window, getParameterByName, $filter, trFilter) {
    var vm = this;
    vm.resetUuid = getParameterByName('reset_uuid', $location.absUrl());
    
    vm.reset = reset;
    vm.resetPassword = resetPassword;
    vm.confirm = confirm;
    vm.cancel = cancel;
    
    vm.hasError = false;
    vm.errorMessage = '';
    
    //Error message dialog handler for resetting password.
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
    
    function reset() {
      vm.hasError = false;
      vm.errorMessage = '';
    }    
        
    function resetPassword(user) {
      if(user && angular.isDefined(user.password)) {

        vm.action = vm.confirm;
        
        console.log('rececived password:' + user.password + ', reset_uuid:' + vm.resetUuid);
        ResetPasswordService(vm.resetUuid, user.password)
          .success(resetPasswordSuccess)
          .error(resetPasswordFailed);
      }
    }
    
    function confirm() {
      $window.location.href = '/';      
    }
    
    function resetPasswordSuccess(data, status) {
      vm.modalTitle = $filter('tr')('reset_password');
      vm.modalMessage = $filter('tr')('successful_reset_password');
      vm.confirmOnly = true;
      $scope.$broadcast('showDialog', true);
    }
    
    function resetPasswordFailed(data) {
      vm.hasError = true;
          
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_reset_pasword') + data);
      $scope.$emit('raiseError', true);
      
      console.log('Failed to reset password:' + data);
    }
    
    function cancel() {
      $window.location.href = '/'; 
    }
  }
  
})();