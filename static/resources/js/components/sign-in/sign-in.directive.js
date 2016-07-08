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
    .module('harbor.sign.in')
    .directive('signIn', signIn);
    
  SignInController.$inject = ['SignInService', 'LogOutService', 'currentUser', 'I18nService', '$window', '$scope', 'getParameterByName', '$location'];
  function SignInController(SignInService, LogOutService, currentUser, I18nService, $window, $scope, getParameterByName, $location) {
    var vm = this;

    vm.hasError = false;
    vm.errorMessage = '';
    
    vm.reset = reset;
    vm.doSignIn = doSignIn;
    vm.doSignUp = doSignUp;
    vm.doForgotPassword = doForgotPassword;
       
    vm.doContinue = doContinue;
    vm.doLogOut = doLogOut;
       
    vm.signInTIP = false;
    
    function reset() {
      vm.hasError = false;
      vm.errorMessage = '';
    } 
      
    function doSignIn(user) {  
      if(user && angular.isDefined(user.principal) && angular.isDefined(user.password)) {
        vm.lastUrl = getParameterByName('last_url', $location.absUrl());
        vm.signInTIP = true;
        SignInService(user.principal, user.password)
          .success(signedInSuccess)
          .error(signedInFailed);
      }
    }
    
    function signedInSuccess(data, status) {
      if(vm.lastUrl) {
        $window.location.href = vm.lastUrl;
        return;
      }
      $window.location.href = "/dashboard";
    }
    
    function signedInFailed(data, status) {
      vm.signInTIP = false;
      if(status === 401) {
        vm.hasError = true;
        vm.errorMessage = 'username_or_password_is_incorrect';
      }
      console.log('Failed to sign in:' + data + ', status:' + status);     
    }
    
    function doSignUp() {
      $window.location.href = '/sign_up';
    }
    
    function doForgotPassword() {
      $window.location.href = '/forgot_password';
    }
    
    function doContinue() {
      $window.location.href = '/dashboard';
    }
    
    function doLogOut() {
      LogOutService()
        .success(logOutSuccess)
        .error(logOutFailed);
    }
    
    function logOutSuccess(data, status) {
      currentUser.unset();
      I18nService().unset();
      $window.location.href= '/';
    }
    
    function logOutFailed(data, status) {
      console.log('Failed to to log out:' + data);
    }
  }
  
  function signIn() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/sign_in',
      'scope': true,
      'controller': SignInController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    
    return directive;

  }
  
})();