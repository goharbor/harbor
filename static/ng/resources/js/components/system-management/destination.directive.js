(function() {
 
  'use strict';
  
  angular
    .module('harbor.system.management')
    .directive('destination', destination);
    
  DestinationController.$inject = ['ListDestinationService'];
  
  function DestinationController(ListDestinationService) {
    var vm = this;
    
    ListDestinationService()
      .then(function(data) {
        vm.destinations = data; 
      }, function() {
                
      });
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