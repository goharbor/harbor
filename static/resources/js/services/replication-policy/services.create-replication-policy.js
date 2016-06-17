(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.replication.policy')
    .factory('CreateReplicationPolicyService', CreateReplicationPolicyService);
    
  CreateReplicationPolicyService.$inject = ['$http'];
  
  function CreateReplicationPolicyService($http) {
    return createReplicationPolicy;
    
    function createReplicationPolicy(policy) {
      return $http      
        .post('/api/policies/replication', {
          'project_id': policy.projectId,
          'target_id': policy.targetId,
          'name': policy.name,
          'enabled': policy.enabled,
          'description': policy.description,
          'cron_str': policy.cronStr,
          'start_time': policy.startTime
        })
    }
  }
    
})();