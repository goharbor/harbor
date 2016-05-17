(function() {
  
  'use strict';
 
   angular
    .module('harbor.services.user')
    .factory('CurrentUserService', CurrentUserService);
  
  CurrentUserService.$inject = ['$http'];
  
  function CurrentUserService($http, $log) {
    
    return CurrentUser;
    
    function CurrentUser() {      
      return $http
        .get('/api/users/current');
    }
  }
})();