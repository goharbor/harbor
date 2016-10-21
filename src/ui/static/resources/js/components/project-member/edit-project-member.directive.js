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
    .directive('editProjectMember', editProjectMember);
      
  EditProjectMemberController.$inject = ['$scope', 'roles', 'getRole','EditProjectMemberService', '$filter', 'trFilter'];
  
  function EditProjectMemberController($scope, roles, getRole, EditProjectMemberService, $filter, trFilter) {
    var vm = this;
        
    vm.roles = roles();
    vm.editMode = false;
    vm.lastRoleName = vm.roleName;
    
    $scope.$watch('vm.roleName', function(current, origin) {
      if(current) {
        vm.currentRole = getRole({'key': 'roleName', 'value': current});  
        vm.roleId = vm.currentRole.id;
      }
    });
    
    vm.updateProjectMember = updateProjectMember;
    vm.deleteProjectMember = deleteProjectMember;
    vm.cancelUpdate = cancelUpdate;
    
    function updateProjectMember(e) {            
      if(vm.editMode) {
        console.log('update project member, roleId:' + e.roleId);         
        EditProjectMemberService(e.projectId, e.userId, e.roleId)
          .success(editProjectMemberComplete)
          .error(editProjectMemberFailed);
      }else {
        vm.editMode = true;      
      } 
    }
    
    function deleteProjectMember() {
      vm.delete();
    }
    
    function editProjectMemberComplete(data, status, headers) {
      console.log('edit project member complete: ' + status);
      vm.lastRoleName = vm.roleName;
      vm.editMode = false;
      vm.reload();      
    }
    
    function editProjectMemberFailed(e) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_change_member'));
      $scope.$emit('raiseError', true);
      console.log('Failed to edit project member:' + e);
    }
    
    function cancelUpdate() {
      vm.editMode = false;
      vm.roleName = vm.lastRoleName;
    }
    
  }
  
  function editProjectMember() {
    var directive = {
      'restrict': 'A',
      'templateUrl': '/static/resources/js/components/project-member/edit-project-member.directive.html',
      'scope': {
        'username': '=',
        'userId': '=',
        'currentUserId': '=',
        'roleName': '=',
        'projectId': '=',
        'delete': '&',
        'reload': '&'
      },
      'controller': EditProjectMemberController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }

})();