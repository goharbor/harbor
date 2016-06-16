(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.replication.job')
    .factory('ListReplicationJobService', ListReplicationJobService);
    
  ListReplicationJobService.$inject = ['$http'];
  
  function ListReplicationJobService($http) {
    
    return listReplicationJob;
    
    function listReplicationJob(policyId, name) {
      return $http
        .get('/api/jobs/replication/', {
          'params': {
            'policy_id': policyId,
            'name': name
          }
        });
    }
  }
})();