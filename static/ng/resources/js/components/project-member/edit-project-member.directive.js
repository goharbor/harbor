(function() {

  'use strict';
  
  angular
    .module('harbor.project.member')
    .directive('editProjectMember', editProjectMember);
      
  EditProjectMemberController.$inject = ['$scope', 'roles', 'getRole','EditProjectMemberService', 'DeleteProjectMemberService'];
  
  function EditProjectMemberController($scope, roles, getRole, EditProjectMemberService, DeleteProjectMemberService) {
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
        vm.editMode = false;
        console.log('update project member, roleId:' + e.roleId);         
        EditProjectMemberService(e.projectId, e.userId, e.roleId)
          .success(editProjectMemberComplete)
          .error(editProjectMemberFailed);
      }else {
        vm.editMode = true;      
      } 
    }
    
    function deleteProjectMember(e) {
      DeleteProjectMemberService(e.projectId, e.userId)
        .success(editProjectMemberComplete)
        .error(editProjectMemberFailed);
      vm.reload();
    }
    
    function editProjectMemberComplete(data, status, headers) {
      console.log('edit project member complete: ' + status);
      vm.lastRoleName = vm.roleName;
    }
    
    function editProjectMemberFailed(e) {
      console.log('Failed to edit project member:' + e);
    }
    
    function cancelUpdate() {
      vm.editMode = false;
      console.log('lastRoleName:' + vm.lastRoleName);
      vm.roleName = vm.lastRoleName;
    }
    
  }
  
  function editProjectMember() {
    var directive = {
      'restrict': 'A',
      'templateUrl': '/static/ng/resources/js/components/project-member/edit-project-member.directive.html',
      'scope': {
        'username': '=',
        'userId': '=',
        'currentUserId': '=',
        'roleName': '=',
        'projectId': '=',
        'reload': '&'
      },
      'controller': EditProjectMemberController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }

})();