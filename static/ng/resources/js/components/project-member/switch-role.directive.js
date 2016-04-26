(function() {
  
  'use strict';
  
  angular
    .module('harbor.project.member')
    .directive('switchRole', switchRole);
  
  SwitchRoleController.$inject = ['getRole', '$scope'];
  
  function SwitchRoleController(getRole, $scope) {
    var vm = this;
    
    vm.currentRole = getRole({'key': 'roleName', 'value': vm.roleName});            
    vm.selectRole = selectRole;
    
    function selectRole(role) {
      vm.currentRole = getRole({'key': 'roleName', 'value': role.roleName});  
      $scope.$emit('changedRoleId', vm.currentRole.id);
    }
    
  }
  
  function switchRole() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/project-member/switch-role.directive.html',
      'scope': {
        'roles': '=',
        'editMode': '=',
        'userId': '=',
        'roleName': '='
      },
      'controller' : SwitchRoleController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  } 
  
})();