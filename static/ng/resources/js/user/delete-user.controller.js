(function() {

  'use strict';
  
  angular
    .module('harbor.user');
    .controller('DeleteUserController', DeleteUserController);
    
  DeleteUserController.$inject = ['DeleteUserService'];
    
  function DeleteUserController(DeleteUserService) {
    
  }
  
})();