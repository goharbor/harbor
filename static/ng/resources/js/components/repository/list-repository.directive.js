(function() {
  'use strict';
  
  angular
    .module('harbor.repository')
    .directive('listRepository', listRepository);   
    
  ListRepositoryController.$inject = ['$scope', '$q', 'ListRepositoryService', 'ListTagService', 'nameFilter'];
  
  function ListRepositoryController($scope, $q, ListRepositoryService, ListTagService, nameFilter) {
    var vm = this;
    
    vm.filterInput = "";
    vm.expand = expand;
        
    vm.retrieve = retrieve;
    
    $scope.$watch('vm.projectId', function(current, origin) {
      if(current) {
         vm.retrieve(current, vm.filterInput);
      }
    });
       
    function retrieve(projectId, filterInput) {
      ListRepositoryService({'projectId': projectId, 'q': filterInput})
        .success(getRepositoryComplete)
        .error(getRepositoryFailed);
    }
   
    function getRepositoryComplete(data, status) {
      console.log(data);
      vm.repositories = data;
    }
    
    function getRepositoryFailed(repsonse) {
      
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
      scope: {
        'projectId': '='
      },
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