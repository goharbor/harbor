(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.user')
    .factory('ForgotPasswordService', ForgotPasswordService);
    
  ForgotPasswordService.$inject = ['$http', '$log'];
  
  function ForgotPasswordService($http, $log) {
    
    return ForgotPassword;
    
    function ForgotPassword(user) {
      
    }
    
  }
  
})();