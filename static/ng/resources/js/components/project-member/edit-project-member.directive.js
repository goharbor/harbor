(function() {

  'use strict';
  
  angular
    .module('harbor.project.member')
    .directive('editProjectMember', editProjectMember);
      
  EditProjectMemberController.$inject = ['$scope', 'roles', 'getRole','EditProjectMemberService', 'DeleteProjectMemberService'];
  
  function EditProjectMemberController($scope, roles, getRole, EditProjectMemberService, DeleteProjectMemberService) {
    var vm = this;
    
    
    $scope.$on('changedRoleId', function(e, val) {
      vm.roleId = val;
    });
    
    vm.roles = roles();
    vm.editMode = false;
    vm.updateProjectMember = updateProjectMember;
    vm.deleteProjectMember = deleteProjectMember;
    
    function updateProjectMember(e) {            
      if(vm.editMode) {
        vm.editMode = false;
      
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
      console.log('editProjectMemberComplete: ' + status);
    }
    
    function editProjectMemberFailed(e) {
      console.log('editProjectMemberFailed:' + e);
    }
    
  }
  
  function editProjectMember() {
    var directive = {
      'restrict': 'A',
      'templateUrl': '/static/ng/resources/js/components/project-member/edit-project-member.directive.html',
      'scope': {
        'username': '=',
        'userId': '=',
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