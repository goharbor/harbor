(function() {

  'use strict';
  
  angular
    .module('harbor.log')
    .directive('listLog', listLog);
    
  ListLogController.$inject  = ['$scope','ListLogService', '$routeParams'];
  
  function ListLogController($scope, ListLogService, $routeParams) {
    var vm = this;
    vm.isOpen = false;
       
    vm.beginTimestamp = 0;
    vm.endTimestamp = 0;
    vm.keywords = "";
    vm.username = "";
        
    vm.op = [];
   
    vm.search = search;
    vm.showAdvancedSearch = showAdvancedSearch;
  
    vm.projectId = $routeParams.project_id;
    vm.queryParams = {
      'beginTimestamp' : vm.beginTimestamp,
      'endTimestamp'   : vm.endTimestamp,
      'keywords' : vm.keywords,
      'projectId': vm.projectId,
      'username' : vm.username
    };
    retrieve(vm.queryParams);
     
    function search(e) {
      if(e.op[0] == 'all') {
        vm.queryParams.keywords = '';
      }else {
        vm.queryParams.keywords = e.op.join('/') ;
      }
      vm.queryParams.username = e.username;
      
      vm.queryParams.beginTimestamp = toUTCSeconds(vm.fromDate, 0, 0, 0);
      vm.queryParams.endTimestamp = toUTCSeconds(vm.toDate, 23, 59, 59);
     
      retrieve(vm.queryParams);
    }
    
    function showAdvancedSearch() {
      if(vm.isOpen){
        vm.isOpen = false;
      }else{
        vm.isOpen = true;
      }
    }
    
    function retrieve(queryParams) {
      ListLogService(queryParams)
        .then(listLogComplete)
        .catch(listLogFailed);
    }

    function listLogComplete(response) {
      vm.logs = response.data;
    }
    function listLogFailed(e){
      console.log('listLogFailed:' + e);
    }
    
    	function toUTCSeconds(date, hour, min, sec) {
      if(date == "") {
        return 0;
      }
      
			var t = new Date(date);
			t.setHours(hour);
			t.setMinutes(min);
			t.setSeconds(sec);
			var utcTime = new Date(t.getUTCFullYear(),
				t.getUTCMonth(), 
				t.getUTCDate(),
				t.getUTCHours(),
				t.getUTCMinutes(),
		    	t.getUTCSeconds());
			return utcTime.getTime() / 1000;
		}
    
  }
  
  function listLog() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/ng/resources/js/components/log/list-log.directive.html',
      replace: true,
      scope: true,
      controller: ListLogController,
      controllerAs: 'vm',
      bindToController: true
    }
    
    return directive;
  }
  
})();