(function() {
  
  'use strict';
  
  angular
    .module('harbor.sign.in')
    .directive('signIn', signIn);
    
  SignInController.$inject = ['SignInService', 'LogOutService', 'currentUser', 'I18nService', '$window', '$scope'];
  function SignInController(SignInService, LogOutService, currentUser, I18nService, $window, $scope) {
    var vm = this;

    vm.hasError = false;
    vm.errorMessage = '';
    
    vm.reset = reset;
    vm.doSignIn = doSignIn;
    vm.doSignUp = doSignUp;
    vm.doForgotPassword = doForgotPassword;
       
    vm.doContinue = doContinue;
    vm.doLogOut = doLogOut;
    
    function reset() {
      vm.hasError = false;
      vm.errorMessage = '';
    } 
      
    function doSignIn(user) {  
      if(user && angular.isDefined(user.principal) && angular.isDefined(user.password)) {
        SignInService(user.principal, user.password)
          .success(signedInSuccess)
          .error(signedInFailed);
      }
    }
    
    function signedInSuccess(data, status) {
      $window.location.href = "/ng/dashboard";
    }
    
    function signedInFailed(data, status) {
      if(status === 401) {
        vm.hasError = true;
        vm.errorMessage = 'username_or_password_is_incorrect';
      }
      console.log('Failed sign in:' + data + ', status:' + status);     
    }
    
    function doSignUp() {
      $window.location.href = '/ng/sign_up';
    }
    
    function doForgotPassword() {
      $window.location.href = '/ng/forgot_password';
    }
    
    function doContinue() {
      $window.location.href = '/ng/dashboard';
    }
    
    function doLogOut() {
      LogOutService()
        .success(logOutSuccess)
        .error(logOutFailed);
    }
    
    function logOutSuccess(data, status) {
      currentUser.unset();
      I18nService().unset();
      $window.location.href= '/ng';
    }
    
    function logOutFailed(data, status) {
      console.log('Failed to log out:' + data);
    }
  }
  
  function signIn() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/ng/sign_in',
      'scope': true,
      'controller': SignInController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    
    return directive;

  }
  
})();