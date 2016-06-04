(function() {
 
  'use strict';
  
  angular
    .module('harbor.system.management')
    .directive('destination', destination);
    
  DestinationController.$inject = ['ListDestinationService'];
  
  function DestinationController(ListDestinationService) {
    var vm = this;
    
    vm.retrieve = retrieve;
    vm.search = search;
    vm.retrieve();
    
    function retrieve() {
      ListDestinationService()
        .success(listDestinationSuccess)
        .error(listDestinationFailed);
    }
    
    function search() {
      vm.retrieve();
    }
    
    function listDestinationSuccess(data, status) {
      vm.destinations = data;
    }
    
    function listDestinationFailed(data, status) {
      console.log('Failed list destination:' + data);
    }
    
    
  }
  
  function destination() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/system-management/destination.directive.html',
      'scope': true,
      'controller': DestinationController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();