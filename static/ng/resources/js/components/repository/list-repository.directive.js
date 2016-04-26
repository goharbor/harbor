(function() {
  'use strict';
  
  angular
    .module('harbor.repository')
    .directive('listRepository', listRepository);   
    
  ListRepositoryController.$inject = ['ListRepositoryService', 'ListTagService', 'nameFilter', '$routeParams'];
  
  function ListRepositoryController(ListRepositoryService, ListTagService, nameFilter, $routeParams) {
    var vm = this;
    
    vm.filterInput = "";
    
    ListRepositoryService({'projectId': $routeParams.project_id, 'q': ''})
      .then(getRepositoryComplete)
      .catch(getRepositoryFailed);
    
    function getRepositoryComplete(response) {
      vm.repositories = response.data;
    }
    
    function getRepositoryFailed(repsonse) {
      
    }
    
    vm.expand = expand;
    
    function expand(e) {
      vm.tags = [];
      ListTagService(e.repoName)
        .then(getTagComplete)
        .catch(getTagFailed);
       
      function getTagComplete(response) {
        vm.tags = response.data;
      }
      
      function getTagFailed(response) {
        
      }
    }
  }
  
  function listRepository() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/ng/resources/js/components/repository/list-repository.directive.html',
      replace: true,
      scope: {
        'info': '='
      },
      controller: ListRepositoryController,
      controllerAs: 'vm',
      bindToController: true
    }
    
    return directive;
  }
  
})();