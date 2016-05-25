(function() {
  
  'use strict';
  
  angular
    .module('harbor.replication')
    .directive('listReplication', listReplication);
    
  ListReplicationController.$inject = [];
  
  function ListReplicationController() {
    var vm = this;
    vm.addReplication = addReplication;
    
    function addReplication() {
      vm.modalTitle = 'Create New Policy';
      vm.modalMessage = '';
    }
    
  }
  
  function listReplication() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/replication/list-replication.directive.html',
      'scope': true,
      'controller': ListReplicationController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();