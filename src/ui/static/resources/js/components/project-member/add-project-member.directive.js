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
    .module('harbor.project.member')
    .directive('addProjectMember', addProjectMember);
   
  AddProjectMemberController.$inject = ['$scope', 'roles', 'AddProjectMemberService'];
 
  function AddProjectMemberController($scope, roles, AddProjectMemberService) {
    var vm = this;
    
    $scope.pm = {};
    
    var pm = $scope.pm;
    
    vm.roles = roles();
    vm.optRole = 1;
   
    vm.save = save;
    vm.cancel = cancel;
    vm.reset = reset;

    vm.hasError = false;
    vm.errorMessage = '';
    
    function save(pm) {     
      if(pm && angular.isDefined(pm.username)) {
        AddProjectMemberService(vm.projectId, vm.optRole, pm.username)
          .success(addProjectMemberComplete)
          .error(addProjectMemberFailed);        
      }
    }    
   
    function cancel(form) {
      
      form.$setPristine();
      form.$setUntouched();
      
      vm.isOpen = false;  
      pm.username = '';
      vm.optRole = 1;
      
      vm.hasError = false;
      vm.errorMessage = '';
    }
        
    function addProjectMemberComplete(data, status, header) {
      console.log('addProjectMemberComplete: status:' + status + ', data:' + data);
      vm.reload();
      vm.isOpen = false;
    }
    
    function addProjectMemberFailed(data, status, headers) {
      if(status === 403) {
        vm.hasError = true;
        vm.errorMessage = 'failed_to_add_member';
      }
      if(status === 409 && pm.username !== '') {
        vm.hasError = true;
        vm.errorMessage = 'username_already_exist';
      }
      if(status === 404) {
        vm.hasError = true;
        vm.errorMessage = 'username_does_not_exist';
      }
      console.log('addProjectMemberFailed: status:' + status + ', data:' + data);
    }
    
    function reset() {
      vm.hasError = false;
      vm.errorMessage = '';
    }
    
  }
  
  function addProjectMember() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/project-member/add-project-member.directive.html',
      'scope': {
        'projectId': '@',
        'isOpen': '=',
        'reload': '&'
      },
      'link': link,
      'controller': AddProjectMemberController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      scope.form.$setPristine();
      scope.form.$setUntouched();
    }
  }
  
})();
