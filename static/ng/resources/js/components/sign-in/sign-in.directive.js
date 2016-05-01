(function() {
  
  'use strict';
  
  angular
    .module('harbor.sign.in')
    .directive('signIn', signIn);
    
  SignInController.$inject = ['SignInService', '$window'];
  function SignInController(SignInService, $window) {
    var vm = this;
    vm.principal = "";
    vm.password = "";
    vm.doSignIn = doSignIn;
 
    function doSignIn() {
      if(vm.principal != "" && vm.password != "") {
        SignInService(vm.principal, vm.password)
          .success(signedInSuccess)
          .error(signedInFailed);
      }else{
        $window.alert('Please input your username or password!');
      }
    }
    
    function signedInSuccess(data, status) {
      console.log(status);
      $window.location.href = "/ng/project";
    }
    
    function signedInFailed(data, status) {
      console.log(status);
    }
    
  }
  
  function signIn() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/sign-in/sign-in.directive.html',
      'controller': SignInController,
      'controllerAs': 'vm',
      'bindToController': true
    }
    return directive;
  }
  
})();