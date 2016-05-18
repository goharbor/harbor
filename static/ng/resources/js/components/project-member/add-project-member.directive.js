(function() {
  
  'use strict';
  
  angular
    .module('harbor.project.member')
    .directive('addProjectMember', addProjectMember);
   
  AddProjectMemberController.$inject = ['$scope', 'roles', 'AddProjectMemberService'];
 
  function AddProjectMemberController($scope, roles, AddProjectMemberService) {
    var vm = this;
    vm.username = '';
    vm.roles = roles();
    vm.optRole = 1;
    
    vm.reset = reset;
    vm.save = save;
    vm.cancel = cancel;

    vm.hasError = false;
    vm.errorMessage = '';
    
    function reset() {
      vm.hasError = false;
      vm.errorMessage = '';
    }
    
    function save(pm) {     
      if(pm && angular.isDefined(pm.username)) {
        AddProjectMemberService(vm.projectId, vm.optRole, pm.username)
          .success(addProjectMemberComplete)
          .error(addProjectMemberFailed);
        vm.username = '';
        vm.optRole = 1;
        vm.reload();
      }
    }    
   
    function cancel(form) {
      if(form) {
        form.$setPristine();
      }
      vm.isOpen = false;  
      vm.username = '';
      vm.optRole = 1;
    }
    
    function addProjectMemberComplete(data, status, header) {
      console.log('addProjectMemberComplete: status:' + status + ', data:' + data);
    }
    
    function addProjectMemberFailed(data, status, headers) {
      if(status === 409) {
        vm.hasError = true;
        vm.errorMessage = 'username_already_exist';
      }
      if(status == 404) {
        vm.hasError = true;
        vm.errorMessage = 'username_does_not_exist';
      }
      console.log('addProjectMemberFailed: status:' + status + ', data:' + data);
    }
    
  }
  
  function addProjectMember() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/project-member/add-project-member.directive.html',
      'scope': {
        'projectId': '@',
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