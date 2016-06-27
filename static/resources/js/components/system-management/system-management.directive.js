(function() {
  
  'use strict';
  
  angular
    .module('harbor.system.management')
    .directive('systemManagement', systemManagement);
    
  function SystemManagementController() {
    var vm = this;
    vm.target = 'destinations';
  }
  
  function systemManagement() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/system-management/system-management.directive.html',
      'scope': true,
      'controller': SystemManagementController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();