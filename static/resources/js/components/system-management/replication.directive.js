(function() {
  
  'use strict';
  
  angular
    .module('harbor.system.management')
    .directive('replication', replication);
  
  ReplicationController.$inject = ['$scope', 'ListReplicationPolicyService', 'ToggleReplicationPolicyService'];
  
  function ReplicationController($scope, ListReplicationPolicyService, ToggleReplicationPolicyService) {
    
    $scope.subsSubPane = 276;
    
    var vm = this;
    vm.retrieve = retrieve;
    vm.search = search;
    vm.togglePolicy = togglePolicy;
    vm.editReplication = editReplication;
    vm.retrieve();
    
    function search() {
      vm.retrieve();
    }
    
    function retrieve() {
      ListReplicationPolicyService('', '', vm.replicationName)
        .success(listReplicationPolicySuccess)
        .error(listReplicationPolicyFailed);
    }
    
    function listReplicationPolicySuccess(data, status) {
      vm.replications = data || [];
    }
    
    function listReplicationPolicyFailed(data, status) {
      console.log('Failed list replication policy.');
    }
    
    function togglePolicy(policyId, enabled) {
      ToggleReplicationPolicyService(policyId, enabled)
        .success(toggleReplicationPolicySuccess)
        .error(toggleReplicationPolicyFailed);
    }
    
    function toggleReplicationPolicySuccess(data, status) {
      console.log('Successful toggle replication policy.');
      vm.retrieve();
    }
    
    function toggleReplicationPolicyFailed(data, status) {
      console.log('Failed toggle replication policy.');
    }
    
    function editReplication(policyId) {
      vm.action = 'EDIT';
      vm.modalTitle = 'Edit policy';
      vm.policyId = policyId;
    }
  }
  
  function replication() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/system-management/replication.directive.html',
      'scope': true,
      'controller': ReplicationController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();