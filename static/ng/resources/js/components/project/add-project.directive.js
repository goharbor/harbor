(function() {
  
  'use strict';
  
  angular
    .module('harbor.project')
    .directive('addProject', addProject);
    
  AddProjectController.$inject = ['AddProjectService', '$scope'];
  
  function AddProjectController(AddProjectService, $scope) {
    var vm = this;
    vm.projectName = "";
    vm.isPublic = false;
    
    vm.addProject = addProject;
    vm.cancel = cancel;
    
    function addProject() {
      
      if(vm.projectName == "") {
        alert("Please input the project name.");
        return;
      }
 
      AddProjectService(vm.projectName, vm.isPublic)
        .success(addProjectSuccess)
        .error(addProjectFailed);
    }
    
    function addProjectSuccess(data, status) {
      vm.isOpen = false;
      vm.projectName = "";
      vm.isPublic = false;
      $scope.$emit('addedSuccess', true);
    }
    
    function addProjectFailed(data, status) {
      console.log('Failed to add project:' + status);
    }
    
    function cancel(){
      vm.isOpen = false;
      vm.projectName = "";
      vm.isPublic = false;
    }
  }
  
  function addProject() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/project/add-project.directive.html',
      'controller': AddProjectController,
      'scope' : {
        'isOpen': '='
      },
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
   
})();