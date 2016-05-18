(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.user')
    .factory('SendMailService', SendMailService);
    
  SendMailService.$inject = ['$http', '$log'];
  
  function SendMailService($http, $log) {
    
    return SendMail;
    
    function SendMail(email) {
      return $http
        .get('/ng/sendEmail', {
          'params': {
            'email': email
          }
        });
    }
    
  }
  
})();