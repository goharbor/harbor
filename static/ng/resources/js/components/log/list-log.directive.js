(function() {

  'use strict';
  
  angular
    .module('harbor.log')
    .directive('listLog', listLog);
    
//  ListLogController.$inject  = ['ListLogService'];
  
  function ListLogController() {
    
  }
  
  function listLog() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/ng/resources/js/components/log/list-log.directive.html',
      replace: true,
      controller: ListLogController,
      controllerAs: 'vm',
      bindToController: true
    }
    
    return directive;
  }
  
})();