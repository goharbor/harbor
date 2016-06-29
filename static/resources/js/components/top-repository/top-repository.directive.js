(function() {
    
    'use strict';
    
    angular
      .module('harbor.top.repository')
      .directive('topRepository', topRepository);
      
    TopRepositoryController.$inject = ['ListTopRepositoryService'];
    
    function TopRepositoryController(ListTopRepositoryService) {
        var vm = this;
        
        ListTopRepositoryService(5)
          .success(listTopRepositorySuccess)
          .error(listTopRepositoryFailed);
          
        function listTopRepositorySuccess(data) {
            vm.top10Repositories = data || []
            console.log(vm.top10Repositories.length);
        }
   
        function listTopRepositoryFailed(data, status) {
            console.log('Failed list integrated logs:' + status);
        }
    }
    
    function topRepository() {
        var directive = {
          'restrict': 'E',
          'templateUrl': '/static/resources/js/components/top-repository/top-repository.directive.html',
          'controller': TopRepositoryController,
          'scope' : {
            'customBodyHeight': '='
          },
          'controllerAs': 'vm',
          'bindToController': true
        };
        
        return directive;
    }
    
})();
