/*
    Copyright (c) 2016 VMware, Inc. All Rights Reserved.
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
        
        http://www.apache.org/licenses/LICENSE-2.0
        
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/
(function() {
  
  'use strict';
  
  angular
    .module('harbor.project')
    .directive('addProject', addProject);
    
  AddProjectController.$inject = ['AddProjectService', '$scope'];
  
  function AddProjectController(AddProjectService, $scope) {
    var vm = this;
    
    $scope.p = {};
    var vm0 = $scope.p;
    vm0.projectName = '';
    vm.isPublic = 0;
    
    vm.addProject = addProject;
    vm.cancel = cancel;
        
    vm.reset = reset;
    
    vm.hasError = false;
    vm.errorMessage = '';
    
    $scope.$watch('vm.isOpen', function(current) {
      if(current) {
        $scope.form.$setPristine();
        $scope.form.$setUntouched();
        vm0.projectName = '';
        vm.isPublic = 0;
      }
    });
    
        
    function addProject(p) {
      if(p && angular.isDefined(p.projectName)) {
        vm.isPublic = vm.isPublic ? 1 : 0;
        AddProjectService(p.projectName, vm.isPublic)
          .success(addProjectSuccess)
          .error(addProjectFailed);
      }
    }
    
    function addProjectSuccess(data, status) {
      $scope.$emit('addedSuccess', true);
      vm.hasError = false;
      vm.errorMessage = '';
      vm.isOpen = false;
    }
    
    function addProjectFailed(data, status) {
      vm.hasError = true;
      if(status === 400 && vm0.projectName !== '' && vm0.projectName.length < 4) {
        vm.errorMessage = 'project_name_is_too_short';
      }
      if(status === 400 && vm0.projectName.length > 30) {
        vm.errorMessage = 'project_name_is_too_long';
      }
      if(status === 409 && vm0.projectName !== '') {
        vm.errorMessage = 'project_already_exist';
      }
      console.log('Failed to add project:' + status);
    }
    
    function cancel(form){
      if(form) {
        form.$setPristine();
        form.$setUntouched();
      }
      vm.isOpen = false;
      vm0.projectName = '';
      vm.isPublic = 0;
    
      vm.hasError = false; 
      vm.errorMessage = '';
    }
   
    function reset() {
      vm.hasError = false;
      vm.errorMessage = '';
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
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;    
  }
   
})();
