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
      
      var projectId = queryParams.projectId;
      var q = queryParams.q;
     
      return $http
        .get('/api/repositories', {
          'params':{
            'project_id': projectId,
            'q': q
          }
      });
    }
  }
})();