(function() {
  'use strict';
  
  angular
    .module('harbor.repository')
    .directive('listRepository', listRepository);   
    
  ListRepositoryController.$inject = ['$scope', 'ListRepositoryService', 'ListTagService', 'nameFilter', '$routeParams'];
  
  function ListRepositoryController($scope, ListRepositoryService, ListTagService, nameFilter, $routeParams) {
    var vm = this;
        
    vm.filterInput = "";
    vm.retrieve = retrieve;
    vm.expand = expand;
    vm.projectId = $routeParams.project_id;
 
    vm.retrieve();
    
    function retrieve(){
      ListRepositoryService(vm.projectId, vm.filterInput)
        .success(getRepositoryComplete)
        .error(getRepositoryFailed);
    }
   
    function getRepositoryComplete(data, status) {
      vm.repositories = data;
    }
    
    function getRepositoryFailed(repsonse) {
      console.log('failed to list repositories:' + response);      
    }
        
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
      link: 'link',
      controller: ListRepositoryController,
      controllerAs: 'vm',
      bindToController: true
    }
    
    return directive;
   
    function link(scope, element, attrs, ctrl) {

    }
    
  }
  
})();