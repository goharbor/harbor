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
    vm.projectId = $routeParams.project_id;
    
    vm.retrieve();

    $scope.$on('repoName', function(e, val) {
      vm.repoName = val;
    });

    $scope.$on('tag', function(e, val){
      vm.tag = val;
    });
    
    vm.message = "Are you sure to delete the tag of image?";
    vm.deleteImage = deleteImage;


    
    function retrieve(){
      ListRepositoryService(vm.projectId, vm.filterInput)
        .success(getRepositoryComplete)
        .error(getRepositoryFailed);
    }
   
    function getRepositoryComplete(data, status) {
      vm.repositories = data;
    }
    
    function getRepositoryFailed(repsonse) {
      console.log('Failed list repositories:' + response);      
    }
   
    function deleteImage() {
      console.log('repoName:' + vm.repoName + ', tag:' + vm.tag);
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