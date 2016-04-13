(function() {

  'use strict';
  
  angular
    .module('harbor.user');
    .controller('SignInController', SignInController);
  
  SignInController.$inject = ['SignInService'];
  
  function SignInController(SignInService) {
    
  }
  
})();