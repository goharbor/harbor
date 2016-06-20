(function() {
 
  'use strict';
  
  angular
    .module('harbor.system.management')
    .directive('destination', destination);
    
  DestinationController.$inject = ['$scope', 'ListDestinationService', 'DeleteDestinationService'];
  
  function DestinationController($scope, ListDestinationService, DeleteDestinationService) {
    var vm = this;
    
    vm.retrieve = retrieve;
    vm.search = search;
    vm.addDestination = addDestination;
    vm.editDestination = editDestination;
    vm.confirmToDelete = confirmToDelete;
    vm.deleteDestination = deleteDestination;
    
    vm.retrieve();
    
    function retrieve() {
      ListDestinationService('', vm.destinationName)
        .success(listDestinationSuccess)
        .error(listDestinationFailed);
    }
    
    function search() {
      vm.retrieve();
    }
    
    function addDestination() {
      vm.action = 'ADD_NEW';
      console.log('Action for destination:' + vm.action);
    }
    
    function editDestination(targetId) {
      vm.action = 'EDIT';
      vm.targetId = targetId;
      console.log('Action for destination:' + vm.action + ', target ID:' + vm.targetId);
    }
    
    function confirmToDelete(targetId) {
      vm.selectedTargetId = targetId;
    }
    
    function deleteDestination() {
      DeleteDestinationService(vm.selectedTargetId)
        .success(deleteDestinationSuccess)
        .error(deleteDestinationFailed);
    }
    
    function listDestinationSuccess(data, status) {
      vm.destinations = data;
    }
    
    function listDestinationFailed(data, status) {
      console.log('Failed list destination:' + data);
    }
    
    function deleteDestinationSuccess(data, status) {
      console.log('Successful delete destination.');
      vm.retrieve();
    }
    
    function deleteDestinationFailed(data, status) {
      console.log('Failed delete destination.');
    }   
  }
  
  function destination() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/system-management/destination.directive.html',
      'scope': true,
      'controller': DestinationController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();