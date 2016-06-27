(function() {
  'use strict';
  
  angular
    .module('harbor.repository')
    .directive('listRepository', listRepository);   
    
  ListRepositoryController.$inject = ['$scope', 'ListRepositoryService', 'DeleteRepositoryService', '$filter', 'trFilter', '$location', 'getParameterByName'];
  
  function ListRepositoryController($scope, ListRepositoryService, DeleteRepositoryService, $filter, trFilter, $location, getParameterByName) {
    
    $scope.subsTabPane = 30;
    
    var vm = this;
  
    vm.filterInput = '';

    var hashValue = $location.hash();
    if(hashValue) {
      var slashIndex = hashValue.indexOf('/');
      if(slashIndex >=0) {
        vm.filterInput = hashValue.substring(slashIndex + 1);      
      }else{
        vm.filterInput = hashValue;
      }
    }
        
    vm.retrieve = retrieve;
    vm.tagCount = {};
    
    vm.projectId = getParameterByName('project_id', $location.absUrl());
    vm.retrieve(); 
    
    $scope.$on('$locationChangeSuccess', function() {
      vm.projectId = getParameterByName('project_id', $location.absUrl());
      vm.retrieve();    
    });
    

    $scope.$watch('vm.repositories', function(current) {
      if(current) {
        vm.repositories = current || [];
      }
    });

    $scope.$on('repoName', function(e, val) {
      vm.repoName = val;
    });

    $scope.$on('tag', function(e, val){
      vm.tag = val;
    });
    
    $scope.$on('tagCount', function(e, val) {
      vm.tagCount = val;
    });
    
    $scope.$on('modalTitle', function(e, val) {
      vm.modalTitle = val;
    });
    
    $scope.$on('modalMessage', function(e, val) {
      vm.modalMessage = val;
    });
    
    vm.deleteByRepo = deleteByRepo;
    vm.deleteImage =  deleteImage;

    function retrieve(){
      ListRepositoryService(vm.projectId, vm.filterInput)
        .success(getRepositoryComplete)
        .error(getRepositoryFailed);
    }
   
    function getRepositoryComplete(data, status) {
      vm.repositories = data || [];
    }
    
    function getRepositoryFailed(response) {
      console.log('Failed list repositories:' + response);      
    }
   
  
    function deleteByRepo(repoName) {
      vm.repoName = repoName;
      vm.tag = '';      
      vm.modalTitle = $filter('tr')('alert_delete_repo_title', [repoName]);
      vm.modalMessage = $filter('tr')('alert_delete_repo', [repoName]);
    }
  
    function deleteImage() {
      console.log('repoName:' + vm.repoName + ', tag:' + vm.tag);
      DeleteRepositoryService(vm.repoName, vm.tag)
        .success(deleteRepositorySuccess)
        .error(deleteRepositoryFailed);
    }
    
    function deleteRepositorySuccess(data, status) {
      vm.retrieve();
    }
    
    function deleteRepositoryFailed(data, status) {
      console.log('Failed delete repository:' + data);
    }
    
  }
  
  function listRepository() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/resources/js/components/repository/list-repository.directive.html',
      controller: ListRepositoryController,
      controllerAs: 'vm',
      bindToController: true
    };
    
    return directive;
  
  }
  
})();