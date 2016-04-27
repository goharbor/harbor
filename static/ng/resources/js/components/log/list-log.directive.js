(function() {

  'use strict';
  
  angular
    .module('harbor.log')
    .directive('listLog', listLog);
    
  ListLogController.$inject  = ['ListLogService', '$routeParams'];
  
  function ListLogController(ListLogService, $routeParams) {
    var vm = this;
    vm.isOpen = false;
    vm.projectId = $routeParams.project_id;
    
    vm.beginTimestamp = 0;
    vm.endTimestamp = 0;
    vm.keywords = "";
    vm.username = "";
    
    vm.op = [];
    vm.others = "";
    
           
    vm.search = search;
    vm.aSearch= aSearch;
    
    vm.advancedSearch = advancedSearch;
  
    
    var queryParams = {
      'beginTimestamp' : vm.beginTimestamp,
      'endTimestamp'   : vm.endTimestamp,
      'keywords' : vm.keywords,
      'projectId': vm.projectId,
      'username' : vm.username
    };

    retrieve(queryParams);

    function search(e) {
      queryParams.username = e.username;
      retrieve(queryParams);
    }
    
    function aSearch(e) {
      if(e.op == 'all') {
        queryParams.keywords = '';
      }else {
        queryParams.keywords = e.op.join('/') ;
      }
      if(e.others != "") {
        queryParams.keywords += '/' + e.others;
      }
      queryParams.username = vm.username;
      
      retrieve(queryParams);
    }
    
    function advancedSearch() {
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