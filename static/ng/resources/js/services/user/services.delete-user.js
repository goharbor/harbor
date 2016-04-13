(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.user')
    .factory('DeleteUserService', DeleteUserService);
    
  DeleteUserService.$inject = ['$http', '$log'];
  
  function DeleteUserService($http, $log) {
    
    return DeleteUser;
    
    function DeleteUser(user) {
      
    }
    
  }
  
})();