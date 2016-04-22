(function() {
  
  'use strict';
  
  angular
    .module('harbor.project.member')
    .directive('addProjectMember', addProjectMember);
   
  AddProjectMemberController.$inject = ['roles', 'AddProjectMemberService'];
 
  function AddProjectMemberController(roles, AddProjectMemberService) {
    var vm = this;
    vm.roles = roles();
    vm.optRole = 1;
    vm.save = save;
    vm.cancel = cancel;
    
    function save() {
      console.log(vm.optRole);
    }    
   
    function cancel() {
      vm.isOpen = false;  
    }
    
  }
  
  function addProjectMember() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/project-member/add-project-member.directive.html',
      'scope': {
        'isOpen': '='
      },
      'controller': AddProjectMemberController,
      'controllerAs': 'vm',
      'bindToController': true
    }
    return directive;
  }
  
})();