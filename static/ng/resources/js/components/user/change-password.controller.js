(function() {

  'use strict';
  
  angular
    .module('harbor.user');
    .controller('ChangePasswordController', ChangePasswordController);
    
  ChangePassswordController.$inject = ['ChangePasswordService'];
    
  function ChangePasswordController(ChangePasswordService) {
    
  }
  
})();