(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.replication.job')
    .factory('ListReplicationJobService', ListReplicationJobService);
    
  ListReplicationJobService.$inject = ['$http'];
  
  function ListReplicationJobService($http) {
    
    return listReplicationJob;
    
    function listReplicationJob(policyId, repository, status, startTime, endTime) {
      return $http
        .get('/api/jobs/replication/', {
          'params': {
            'policy_id': policyId,
            'repository': repository,
            'status': status,
            'start_time': startTime,
            'end_time': endTime
          }
        });
    }
  }
})();