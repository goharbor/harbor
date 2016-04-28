(function() {
  
  'use strict';
  
  angular
    .module('harbor.project.member')
    .directive('addProjectMember', addProjectMember);
   
  AddProjectMemberController.$inject = ['roles', 'AddProjectMemberService'];
 
  function AddProjectMemberController(roles, AddProjectMemberService) {
    var vm = this;
    vm.username = "";
    vm.roles = roles();
    vm.optRole = 1;
    vm.save = save;
    vm.cancel = cancel;
    
    function save() {
      
      AddProjectMemberService(2, vm.optRole, vm.username)
        .success(addProjectMemberComplete)
        .error(addProjectMemberFailed);
      vm.isOpen = false;
      vm.username = "";
      vm.optRole = 1;
      vm.reload();
    }    
   
    function cancel() {
      vm.isOpen = false;  
    }
    
    function addProjectMemberComplete(data, status, header) {
      console.log('addProjectMemberComplete: status:' + status + ', data:' + data);
    }
    
    function addProjectMemberFailed(data, status, headers) {
      console.log('addProjectMemberFailed: status:' + status + ', data:' + data);
    }
    
  }
  
  function addProjectMember() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/project-member/add-project-member.directive.html',
      'scope': {
        'isOpen': '=',
        'reload': '&'
      },
      'controller': AddProjectMemberController,
      'controllerAs': 'vm',
      'bindToController': true
    }
    return directive;
  }
  
})();