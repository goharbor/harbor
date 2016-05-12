(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.user')
    .factory('LogOutService', LogOutService);
  
  LogOutService.$inject = ['$http'];
    
  function LogOutService($http) {
    return logOut;
    function logOut() {
      return $http
        .get('/ng/log_out');
    }
  }
  
})();