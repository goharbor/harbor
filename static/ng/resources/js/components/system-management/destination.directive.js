(function() {
 
  'use strict';
  
  angular
    .module('harbor.system.management')
    .directive('destination', destination);
    
  function DestinationController() {
    var vm = this;
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