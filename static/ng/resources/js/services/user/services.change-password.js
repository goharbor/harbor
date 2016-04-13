(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.user')
    .factory('ChangePasswordService', ChangePasswordService);
    
  ChangePasswordService.$inject = ['$http', '$log'];
  
  function ChangePasswordService($http, $log) {
    
    return ChangePassword;
    
    function ChangePassword(user) {
      
    }
    
  }
  
})();