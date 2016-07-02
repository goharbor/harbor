(function() {
  
  'use strict';
  
  angular
    .module('harbor.system.management')
    .directive('systemManagement', systemManagement);

  SystemManagementController.$inject = ['$scope', '$location'];
    
  function SystemManagementController($scope, $location) {
    var vm = this;
    vm.target = 'destinations';
    $scope.$on('$locationChangeSuccess', function(e) {
      vm.target = $location.path().substring(1);
    });
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