(function() {

  'use strict';
  
  angular
    .module('harbor.project.member')
    .constant('roles', roles)
    .directive('editProjectMember', editProjectMember);
  
  function roles() {
    return [
      {'id': '1', 'name': 'Project Admin'},
      {'id': '2', 'name': 'Developer'},
      {'id': '3', 'name': 'Guest'}
    ];
  }
    
  EditProjectMemberController.$inject = ['roles', 'EditProjectMemberService', '$scope'];
  
  function EditProjectMemberController(roles, EditProjectMemberService, $scope) {
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
    }
  }
  
  function editProjectMember() {
    var directive = {
      'restrict': 'A',
      'templateUrl': '/static/ng/resources/js/components/project-member/edit-project-member.directive.html',
      'scope': {
        'username': '=',
        'userId': '=',
        'roleId': '='
      },
      'link': link,
      'controller': EditProjectMemberController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs) {

    }
  }

})();