(function() {

  'use strict';
 
   angular
    .module('harbor.services.user')
    .factory('SignUpService', SignUpService);
  
  SignUpService.$inject = ['$http', '$log'];
  
  function SignUpService($http, $log) {
    
    return SignUp;
    
    function SignUp(user) {      
      return $http
        .post('/api/user', {
          'username': user.username,
          'email': user.email,
          'password': user.password,
          'realname': user.realname,
          'comment': user.comment
        });
    }
  }
})();