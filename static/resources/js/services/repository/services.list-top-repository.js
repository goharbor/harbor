(function() {
  'use strict';
 
   angular
    .module('harbor.services.repository')
    .factory('ListTopRepositoryService', ListTopRepositoryService);
  
  ListTopRepositoryService.$inject = ['$http', '$log'];
  
  function ListTopRepositoryService($http, $log) {
    
    return listTopRepository;
    
    function listTopRepository(count) {
      $log.info('Get public repositories which are accessed most:');
      return $http
        .get('/api/repositories/top', {
          'params' : {
            'count': count,
          }
        });
      
    }
  }
  
})();