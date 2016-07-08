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
    .module('harbor.layout.forgot.password')
    .controller('ForgotPasswordController', ForgotPasswordController);
  
  ForgotPasswordController.$inject = ['SendMailService', '$window', '$scope', '$filter', 'trFilter'];
  
  function ForgotPasswordController(SendMailService, $window, $scope, $filter, trFilter) {
    var vm = this;
    
    vm.hasError = false;
    vm.show = false;
    vm.errorMessage = '';
    
    vm.reset = reset;
    vm.sendMail = sendMail;
    
    vm.confirm = confirm;
    vm.toggleInProgress = false;
        
    //Error message dialog handler for forgotting password.
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
    
    function reset(){
      vm.hasError = false;
      vm.errorMessage = '';
    }
    
    function sendMail(user) {
      if(user && angular.isDefined(user.email)) { 
        
        vm.action = vm.confirm;
        
        vm.toggleInProgress = true;
        SendMailService(user.email)
          .success(sendMailSuccess)
          .error(sendMailFailed);
      }
    }
    
    function sendMailSuccess(data, status) {
      vm.toggleInProgress = false;
      vm.modalTitle = $filter('tr')('forgot_password');
      vm.modalMessage = $filter('tr')('mail_has_been_sent');
      vm.confirmOnly = true;
      $scope.$broadcast('showDialog', true);
    }
    
    function sendMailFailed(data, status) {
      vm.toggleInProgress = false;
      vm.hasError = true;
      vm.errorMessage = data;
      
      if(status === 500) {
        $scope.$emit('modalTitle', $filter('tr')('error'));
        $scope.$emit('modalMessage', $filter('tr')('failed_to_send_email'));        
        $scope.$emit('raiseError', true);
      }
      console.log('Failed to send mail:' + data);
    }
    
    function confirm() {
      $window.location.href = '/';
    }
   
    
  }
  
})();