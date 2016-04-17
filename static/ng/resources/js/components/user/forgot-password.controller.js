(function() {

  'use strict';
  
  angular
    .module('harbor.user');
    .controller('ForgotPasswordController', ForgotPasswordController);
   
  
  ForgotPasswordController.$inject = ['ForgotPasswordService'];
   
  function ForgotPasswordController(ForgotPasswordService) {
    
  }
  
})();