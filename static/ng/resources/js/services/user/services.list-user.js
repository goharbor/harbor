(function() {

  'use strict';
 
   angular
    .module('harbor.services.user')
    .factory('ListUserService', ListUserService);
  
  ListUserService.$inject = ['$http', '$log'];
  
  function ListUserService($http, $log) {
    
    return ListUser;
    
    function ListUser(queryParams) {      
      $log.info(queryParams);
    }
  }
})();