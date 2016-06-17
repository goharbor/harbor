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
    
    vm.reset = reset;
    vm.addProject = addProject;
    vm.cancel = cancel;
    
    vm.hasError = false;
    vm.errorMessage = '';
    
    function reset() {
      vm.hasError = false;
      vm.errorMessage = '';
    }
    
    function addProject(p) {
      if(p && angular.isDefined(p.projectName)) {
        AddProjectService(p.projectName, vm.isPublic)
          .success(addProjectSuccess)
          .error(addProjectFailed);
      }
    }
    
    function addProjectSuccess(data, status) {
      vm.projectName = "";
      vm.isPublic = false;     
      $scope.$emit('addedSuccess', true);
    }
    
    function addProjectFailed(data, status) {
      vm.hasError = true;
      if(status == 400) {
        vm.errorMessage = 'project_name_is_invalid';        
      }
      if(status === 409) {
        vm.errorMessage = 'project_already_exist';
      }
      console.log('Failed to add project:' + status);
    }
    
    function cancel(form){
      if(form) {
        form.$setPristine();
      }
      vm.isOpen = false;
      vm.projectName = '';
      vm.isPublic = false;
    }
  }
  
  function addProject() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/project/add-project.directive.html',
      'controller': AddProjectController,
      'scope' : {
        'isOpen': '='
      },
      'link': link,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;

    function link(scope, element, attrs, ctrl) {
      
    }
  }
   
})();