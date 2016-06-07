(function() {
  
  'use strict';
  
  angular
    .module('harbor.system.management')
    .directive('createDestination', createDestination);
    
  CreateDestinationController.$inject = ['CreateDestinationService'];
  
  function CreateDestinationController(CreateDestinationService) {
    var vm = this;
    vm.save = save;
    
    function save(destination) {
      if(destination) {
        console.log('destination:' + angular.toJson(destination));
        CreateDestinationService(destination.name, destination.endpoint, 
         destination.username, destination.password)
          .success(createDestinationSuccess)
          .error(createDestinationFailed);
      }
    }
    
    function createDestinationSuccess(data, status) {
      alert('Successful created destination.');
    }
    
    function createDestinationFailed(data, status) {
      console.log('Failed create destination:' + data);
    }
    
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