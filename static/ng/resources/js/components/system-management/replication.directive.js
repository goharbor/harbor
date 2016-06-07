(function() {
  
  'use strict';
  
  angular
    .module('harbor.system.management')
    .directive('replication', replication);
  
  ReplicationController.$inject = ['ListReplicationPolicyService'];
  
  function ReplicationController(ListReplicationPolicyService) {
    var vm = this;
    ListReplicationPolicyService()
      .then(function(data) {
        vm.replications = data;
      }, function(data) {
        
      });
  }
  
  function replication() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/system-management/replication.directive.html',
      'scope': true,
      'controller': ReplicationController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();