(function() {

  'use strict';
  
  angular
    .module('harbor.project.member')
    .directive('editProjectMember', editProjectMember);
      
  EditProjectMemberController.$inject = ['roles', 'EditProjectMemberService', 'DeleteProjectMemberService'];
  
  function EditProjectMemberController(roles, EditProjectMemberService) {
    var vm = this;
    vm.roles = roles();
    vm.editMode = false;
    vm.updateProjectMember = updateProjectMember;
    vm.deleteProjectMember = deleteProjectMember;
    
    function updateProjectMember(e) {            
      if(vm.editMode) {
        vm.editMode = false;
               
        console.log('project_id:' + e.projectId + ', user_id:' + e.userId + ', role_id:' + e.roleId);
        
      }else {
        vm.editMode = true;      
      } 
    }
    
    function deleteProjectMember(e) {
      console.log('project_id:' + e.projectId + ', user_id:' + e.userId);
    }
  }
  
  function editProjectMember() {
    var directive = {
      'restrict': 'A',
      'templateUrl': '/static/ng/resources/js/components/project-member/edit-project-member.directive.html',
      'scope': {
        'username': '=',
        'userId': '=',
        'roleId': '=',
        'projectId': '='
      },
      'controller': EditProjectMemberController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }

})();