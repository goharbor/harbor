(function() {
  
  'use strict';
  
  angular
    .module('harbor.system.management')
    .directive('createDestination', createDestination);
    
  CreateDestinationController.$inject = [];
  
  function CreateDestinationController() {
    var vm = this;
  }
  
  function createDestination() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/system-management/create-destination.directive.html',
      'scope': true,
      'controller': CreateDestinationController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();