(function() {

  'use strict';
  
  angular
    .module('harbor.user');
    .controller('SignUpController', SignUpController);
    
  SignUpController.$inject = ['SignUpService'];
  
  function SignUpController(SignUpService) {
    
  }
  
})();