(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.replication.policy')
    .factory('ToggleReplicationPolicyService', ToggleReplicationPolicyService);
    
  ToggleReplicationPolicyService.$inject = ['$http'];  
    
  function ToggleReplicationPolicyService($http) {
    return toggleReplicationPolicy;
    function toggleReplicationPolicy(policyId, enabled) {
      return $http
        .put('/api/policies/replication/' + policyId + '/enablement', {
          'enabled': enabled
        });
    }
  }
  
})();