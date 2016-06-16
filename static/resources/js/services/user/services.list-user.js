(function() {

  'use strict';
 
   angular
    .module('harbor.services.user')
    .factory('ListUserService', ListUserService);
  
  ListUserService.$inject = ['$http', '$log'];
  
  function ListUserService($http, $log) {
    
    return listUser;
    
    function listUser(username) {      
      return $http
        .get('/api/users', {
          'params' : {
            'username': username
          }
        });
    }
  }
})();