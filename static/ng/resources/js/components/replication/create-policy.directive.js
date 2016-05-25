(function() {
  
  'use strict';
  
  angular
    .module('harbor.replication')
    .directive('createPolicy', createPolicy);
  
  function CreatePolicyController() {
    var vm = this;
  }
  
  function createPolicy() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/replication/create-policy.directive.html',
      'scope': true,
      'replace': true,
      'controller': CreatePolicyController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();