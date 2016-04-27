(function() {
  
  'use strict';
  
  angular
    .module('harbor.log')
    .directive('advancedSearch', advancedSearch);
  
  AdvancedSearchController.$inject = ['$scope', 'ListLogService'];
  
  function AdvancedSearchController($scope, ListLogService) {
    var vm = this;
    
    vm.checkOperation = checkOperation;
    
    vm.opAll = true;
    vm.opCreate = true;
    vm.opPull = true;
    vm.opPush = true;
    vm.opDelete = true;
    vm.opOthers = true;
   
    vm.op = [];
    vm.op.push('all');
    function checkOperation(e) {        
      if(e.checked == 'all') {
        vm.opCreate = vm.opAll;
        vm.opPull = vm.opAll;
        vm.opPush = vm.opAll;
        vm.opDelete = vm.opAll;
        vm.opOthers = vm.opAll;
      }else {
        vm.opAll = false;
      }
      
      vm.op = [];
      
      if(vm.opCreate) {
        vm.op.push('create');
      }
      if(vm.opPull) {
         vm.op.push('pull');
      } 
      if(vm.opPush) {
         vm.op.push('push');
      }
      if(vm.opDelete) {
         vm.op.push('delete');
      }
      if(vm.opOthers) {
         vm.op.push(vm.others);
      }
    }
  }
  
  function advancedSearch() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/log/advanced-search.directive.html',
      'scope': {
        'isOpen': '=',
        'op': '=',
        'others': '=',
        'search': '&'
      },
      'controller': AdvancedSearchController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();