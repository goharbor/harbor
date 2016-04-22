(function() {

  'use strict';
  
  angular
    .module('harbor.log')
    .directive('listLog', listLog);
    
  ListLogController.$inject  = ['ListLogService'];
  
  function ListLogController(ListLogService) {
    var vm = this;
    vm.isOpen = false;
    vm.advancedSearch = advancedSearch;
    
    function advancedSearch() {
      if(vm.isOpen){
        vm.isOpen = false;
      }else{
        vm.isOpen = true;
      }
    }
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