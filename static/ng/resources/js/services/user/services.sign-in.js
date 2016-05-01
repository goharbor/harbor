(function() {

  'use strict';
 
   angular
    .module('harbor.services.user')
    .factory('SignInService', SignInService);
  
  SignInService.$inject = ['$http', '$log'];
  
  function SignInService($http, $log) {
    
    return SignIn;
    
    function SignIn(principal, password) {
      return $http({
          method: 'POST',
          url: '/login',
          headers: {'Content-Type': 'application/x-www-form-urlencoded'},
          transformRequest: function(obj) {
              var str = [];
              for(var p in obj)
              str.push(encodeURIComponent(p) + "=" + encodeURIComponent(obj[p]));
              return str.join("&");
          },
          data: {'principal': principal, 'password': password}
      });
    }
  }
})();