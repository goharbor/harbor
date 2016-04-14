(function() {

  'use strict';
  
  angular
    .module('harbor.user');
    .controller('ListUserController', ListUserController);
    
  ListUserController.$inject = ['ListUserService'];
    
  function ListUserController(ListUserService) {
    
  }
  
})();