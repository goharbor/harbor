(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.project')
    .controller('ProjectController', ProjectController);

  ProjectController.$inject = ['$scope', 'ListProjectService', 'CurrentUserService']; 

  function ProjectController($scope, ListProjectService, CurrentUserService) {
    var vm = this;
    
    vm.isOpen = false;
    vm.projectName = '';
    vm.publicity = 0;
    
    vm.retrieve = retrieve;
    vm.getCurrentUser = getCurrentUser;
    vm.showAddProject = showAddProject;
    vm.searchProject = searchProject;    
    vm.showAddButton = showAddButton;
    vm.togglePublicity = togglePublicity;
    
    vm.retrieve();
    
    function retrieve() { 
      $.when(
        CurrentUserService()
          .success(getCurrentUserSuccess)
          .error(getCurrentUserFailed))
      .then(function(){
        ListProjectService(vm.projectName, vm.publicity)
          .success(listProjectSuccess)
          .error(listProjectFailed);
      });
    }
    
    function listProjectSuccess(data, status) {
      vm.projects = data;
    }
    
    function listProjectFailed(e) {
      console.log('Failed to list Project:' + e);
    }
    
    function getCurrentUser() {
      CurrentUserService()
        .success(getCurrentUserSuccess)
        .error(getCurrentUserFailed);
    }
    
    function getCurrentUserSuccess(data, status) {
      vm.currentUser = data;
    }
    
    function getCurrentUserFailed(e) {
      console.log('Failed in getCurrentUser:' + e);
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