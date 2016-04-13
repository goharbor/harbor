(function() {

  'use strict';
 
   angular
    .module('harbor.services.user')
    .factory('SignUpService', SignUpService);
  
  SignUpService.$inject = ['$http', '$log'];
  
  function SignUpService($http, $log) {
    
    return SignUp;
    
    function SignUp(user) {      
      $log.info(user);
    }
  }
})();