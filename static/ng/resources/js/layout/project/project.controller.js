(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.project')
    .controller('ProjectController', ProjectController);

  ProjectController.$inject = ['$scope']; 

  function ProjectController($scope) {
    var vm = $scope;
    vm.showAddProject = showAddProject;
    vm.isOpen = false;
    vm.searchProject = searchProject;
    
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
  }
  
})();