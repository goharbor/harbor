(function() {
  
  'use strict';
  
  angular
    .module('harbor.system.management')
    .directive('systemManagement', systemManagement);

  SystemManagementController.$inject = ['$scope', '$location'];
    
  function SystemManagementController($scope, $location) {
    var vm = this;    
    var currentTarget = $location.path().substring(1);
   
    switch(currentTarget) {
    case 'destinations':
    case 'replication':
      $location.path('/' + currentTarget);
      vm.target = currentTarget;
      break;
    default:
      $location.path('/destinations');
      vm.target = 'destinations';
    }
    
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