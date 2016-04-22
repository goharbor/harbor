(function() {
  
  'use strict';
  
  angular
    .module('harbor.log')
    .directive('advancedSearch', advancedSearch);
  
  AdvancedSearchController.$inject = ['ListLogService'];
  
  function AdvancedSearchController(ListLogService) {
    var vm = this;
    
    vm.checkOperation = checkOperation;
    vm.search = search;
    
    vm.opAll = true;
    vm.opCreate = true;
    vm.opPull = true;
    vm.opPush = true;
    vm.opDelete = true;
    vm.opOthers = true;
    
    function checkOperation(e) {        
      if(e.checked == 'all') {
        vm.opCreate = vm.opAll;
        vm.opPull = vm.opAll;
        vm.opPush = vm.opAll;
        vm.opDelete = vm.opAll;
        vm.opOthers = vm.opAll;
      }
    }
    
    function search() {
      vm.isOpen = false;
    }
  }
  
  function advancedSearch() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/log/advanced-search.directive.html',
      'scope': {
        'isOpen': '='
      },
      'controller': AdvancedSearchController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();