(function() {
  
  'use strict';
  
  angular
    .module('harbor.log')
    .directive('advancedSearch', advancedSearch);
  
  AdvancedSearchController.$inject = ['$scope', 'ListLogService'];
  
  function AdvancedSearchController($scope, ListLogService) {
    var vm = this;
   
    vm.checkOperation = checkOperation;
    vm.close = close;
    
    vm.opAll = true;
    
    vm.doSearch = doSearch;
    
    $scope.$watch('vm.op', function(current) {
      if(current && vm.op[0] === 'all') {
        vm.opCreate = true;
        vm.opPull = true;
        vm.opPush = true;
        vm.opDelete = true;
        vm.opOthers = true;
      }
    }, true);
    
    $scope.$watch('vm.fromDate', function(current) {
      if(current) {
        vm.fromDate = current;
      }
    });
    
    $scope.$watch('vm.toDate', function(current) {
      if(current) {
        vm.toDate = current;
      }
    });
                
               
    vm.opCreate = true;
    vm.opPull = true;
    vm.opPush = true;
    vm.opDelete = true;
    vm.opOthers = true;
    vm.others = '';
         
    vm.op = [];
    vm.op.push('all');
    function checkOperation(e) {        
      if(e.checked === 'all') {
        vm.opCreate = vm.opAll;
        vm.opPull = vm.opAll;
        vm.opPush = vm.opAll;
        vm.opDelete = vm.opAll;
        vm.opOthers = vm.opAll;
      }else {
        vm.op = [];
        vm.opAll = false;
      }
      
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
      if(vm.opOthers && $.trim(vm.others) !== '') {
        vm.op.push($.trim(vm.others));
      }
    }   
    
    vm.pickUp = pickUp;
    
    function pickUp(e) {
      switch(e.key){
      case 'fromDate':
        vm.fromDate = e.value;  
        break;
      case 'toDate':
        vm.toDate = e.value;
        break;
      }
      $scope.$apply();
    }
    
    function close() {
      vm.op = [];
      vm.op.push('all');
      vm.fromDate = '';
      vm.toDate = '';
      vm.others = '';
      vm.isOpen = false;
    }
    
    function doSearch (e){
      if(vm.opOthers && $.trim(vm.others) !== '') {
        e.op.push(vm.others);
      }
      vm.search(e);
    }
  }
  
  function advancedSearch() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/log/advanced-search.directive.html',
      'scope': {
        'isOpen': '=',
        'op': '=',
        'opOthers': '=',
        'others': '=',
        'fromDate': '=',
        'toDate': '=',
        'search': '&'
      },
      'link': link,
      'controller': AdvancedSearchController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      element.find('.datetimepicker').datetimepicker({
				locale: 'en-US',
				ignoreReadonly: true,
				format: 'L',
				showClear: true
		  });      
      element.find('#fromDatePicker').on('blur', function(){
        ctrl.pickUp({'key': 'fromDate', 'value': $(this).val()});
      });
      element.find('#toDatePicker').on('blur', function(){
        ctrl.pickUp({'key': 'toDate', 'value': $(this).val()});
      });
    }
  }
  
})();