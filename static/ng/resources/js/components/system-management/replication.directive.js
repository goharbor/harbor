(function() {
  
  'use strict';
  
  angular
    .module('harbor.system.management')
    .directive('replication', replication);
  
  function ReplicationController() {
    var vm = this;
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