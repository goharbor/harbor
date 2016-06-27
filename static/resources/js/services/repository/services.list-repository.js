(function() {
  'use strict';
 
   angular
    .module('harbor.services.repository')
    .factory('ListRepositoryService', ListRepositoryService);
  
  ListRepositoryService.$inject = ['$http', '$log'];
  
  function ListRepositoryService($http, $log) {
    
    return ListRepository;
    
    function ListRepository(projectId, q) {      
      $log.info('list repositories:' + projectId + ', q:' + q);
  
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