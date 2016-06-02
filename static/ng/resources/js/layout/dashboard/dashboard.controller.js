(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.dashboard')
    .controller('DashboardController', DashboardController);
  
  DashboardController.$inject = ['ListTop10RepositoryService', 'ListIntegratedLogService'];
  
  function DashboardController(ListTop10RepositoryService, ListIntegratedLogService) {
    var vm = this;
  
    ListTop10RepositoryService()
      .then(listTop10RepositorySuccess, listTop10RepositoryFailed);
     
    ListIntegratedLogService()
      .then(listIntegratedLogSuccess, listIntegratedLogFailed);
     
    function listTop10RepositorySuccess(data) {
      vm.top10Repositories = data;
    }
    
    function listTop10RepositoryFailed(data) {
      console.log('Failed list top 10 repositories:' + data);
    }
    
    function listIntegratedLogSuccess(data) {
      vm.integratedLogs = data;
    }
    
    function listIntegratedLogFailed(data) {
      console.log('Failed list integrated logs:' + data);
    }
    
  }
  
})();