(function() {
  'use strict';
 
   angular
    .module('harbor.services.repository')
    .factory('ListRepositoryService', ListRepositoryService);
  
  ListRepositoryService.$inject = ['$http', '$log'];
  
  function ListRepositoryService($http, $log) {
    
    return RepositoryResult;
    
    function RepositoryResult(queryParams) {      
      $log.info(queryParams);
    }
  }
})();