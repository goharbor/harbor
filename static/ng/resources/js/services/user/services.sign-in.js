(function() {

  'use strict';
 
   angular
    .module('harbor.services.user')
    .factory('SignInService', SignInService);
  
  SignInService.$inject = ['$http', '$log'];
  
  function SignInService($http, $log) {
    
    return SignIn;
    
    function SignIn(user) {      
      $log.info(user);
    }
  }
})();