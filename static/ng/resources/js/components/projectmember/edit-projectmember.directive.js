(function() {

  'use strict';
  
  angular
    .module('harbor.projectmember')
    .constant('roles', roles)
    .directive('editProjectMember', editProjectMember);
  
  function roles() {
    return [
      {'id': '1', 'name': 'Project Admin'},
      {'id': '2', 'name': 'Developer'},
      {'id': '3', 'name': 'Guest'}
    ];
  }
    
  EditProjectMemberController.$inject = ['roles', 'EditProjectMemberService'];
  
  function EditProjectMemberController(roles, EditProjectMemberService) {
    var vm = this;
    vm.roles = roles();
    vm.editMode = false;
    vm.update = update;
        
    function update(e) {
      if(vm.editMode) {
        vm.editMode = false;
      }else {
        vm.editMode = true;
      }
      vm.roleId = e.roleId;
    }
    
  }
  
  function editProjectMember() {
    var directive = {
      'restrict': 'A',
      'templateUrl': '/static/ng/resources/js/components/projectmember/edit-projectmember.directive.html',
      'scope': {
        'userId': '=',
        'roleId': '='
      },
      'controller': EditProjectMemberController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }

})();