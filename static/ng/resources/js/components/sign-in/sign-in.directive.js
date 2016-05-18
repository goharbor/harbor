(function() {
  
  'use strict';
  
  angular
    .module('harbor.sign.in')
    .directive('signIn', signIn);
    
  SignInController.$inject = ['SignInService', '$window', '$scope'];
  function SignInController(SignInService, $window, $scope) {
    var vm = this;

    vm.hasError = false;
    vm.errorMessage = '';
    
    vm.reset = reset;
    vm.doSignIn = doSignIn;
    vm.doSignUp = doSignUp;
    vm.doForgotPassword = doForgotPassword;
       
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
      $window.location.href = "/ng/project";
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
  }
  
  function signIn() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/sign-in/sign-in.directive.html',
      'scope': true,
      'controller': SignInController,
      'controllerAs': 'vm',
      'bindToController': true
    }
    return directive;

  }
  
})();