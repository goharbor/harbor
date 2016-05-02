(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.project')
    .controller('ProjectController', ProjectController);

  ProjectController.$inject = ['$scope']; 

  function ProjectController($scope) {
    var vm = this;
    vm.showAddProject = showAddProject;
    vm.isOpen = false;
    vm.searchProject = searchProject;
    vm.inputProjectName = "";
    vm.inputPublicity = 0;
    
    vm.showAddButton = showAddButton;
    vm.togglePublicity = togglePublicity;
    
    $scope.$on('addedSuccess', function(e, val) {
      $scope.$broadcast('needToReload', true);
    });
    
    function showAddProject() {
      if(vm.isOpen){
        vm.isOpen = false;
      }else{
        vm.isOpen = true;
      }
    }
    
    function searchProject() {
      $scope.$broadcast('needToReload', true);
    }
    
    function showAddButton() {
      if(vm.inputPublicity == 0) {
        return true;
      }else{
        return false;
      }
    }
    
    function togglePublicity(e) {
      vm.inputPublicity = e.publicity;
      $scope.$broadcast('needToReload', true);
      console.log('vm.inputPublicity:' + vm.inputPublicity);
    }
    
  }
  
})();