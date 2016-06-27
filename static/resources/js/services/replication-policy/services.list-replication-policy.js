(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.replication.policy')
    .factory('ListReplicationPolicyService', ListReplicationPolicyService);
    
  ListReplicationPolicyService.$inject = ['$http'];
  
  function ListReplicationPolicyService($http) {
       
    return listReplicationPolicy;
    
    function listReplicationPolicy(policyId, projectId, name) {
      return $http
        .get('/api/policies/replication/' + policyId, {
          'params': {
            'project_id': projectId,
            'name': name
          }
        });
    }
    
  }
  
})();