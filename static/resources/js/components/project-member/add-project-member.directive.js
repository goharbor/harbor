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
    }
    
    function addProjectMemberFailed(data, status, headers) {
      if(status === 403) {
        vm.hasError = true;
        vm.errorMessage = 'failed_to_add_member';
      }
      if(status === 409 && pm.username != '') {
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