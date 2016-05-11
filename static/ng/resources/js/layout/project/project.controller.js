(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.project')
    .controller('ProjectController', ProjectController);

  ProjectController.$inject = ['$scope', 'ListProjectService']; 

  function ProjectController($scope, ListProjectService) {
    var vm = this;
    
    vm.isOpen = false;
    vm.projectName = '';
    vm.publicity = 0;
     
    vm.retrieve = retrieve;
    vm.showAddProject = showAddProject;
    vm.searchProject = searchProject;    
    vm.showAddButton = showAddButton;
    vm.togglePublicity = togglePublicity;
    
    $scope.$on('currentUser', function(e, val) {
      vm.currentUser = val;
    });
    
    vm.retrieve();
    
    function retrieve() {       
       
      ListProjectService(vm.projectName, vm.publicity)
        .success(listProjectSuccess)
        .error(listProjectFailed);
    }
    
    function listProjectSuccess(data, status) {
      vm.projects = data;
    }
    
    function listProjectFailed(e) {
      console.log('Failed to list Project:' + e);
    }
          
    $scope.$on('addedSuccess', function(e, val) {
      vm.retrieve();
    });
    
    function showAddProject() {
      if(vm.isOpen){
        vm.isOpen = false;
      }else{
        vm.isOpen = true;
      }
    }
    
    function searchProject() {
      vm.retrieve();
    }
    
    function showAddButton() {
      if(vm.publicity == 0) {
        return true;
      }else{
        return false;
      }
    }
    
    function togglePublicity(e) {
      vm.publicity = e.publicity;
      vm.retrieve();
      console.log('vm.publicity:' + vm.publicity);
    }
    
  }
  
})();