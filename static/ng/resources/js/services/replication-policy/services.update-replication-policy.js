(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.replication.policy')
    .factory('UpdateReplicationPolicyService', UpdateReplicationPolicyService);
    
  UpdateReplicationPolicyService.$inject = ['$http'];
  
  function UpdateReplicationPolicyService($http) {
    return updateReplicationPolicy;
    function updateReplicationPolicy(policyId, policy) {
      return $http
        .put('/api/policies/replication/' + policyId, {
          'name': policy.name,
          'description': policy.description,
          'enabled': policy.enabled
        });
    }
  } 
  
})();