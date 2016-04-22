(function() {
  
  'use strict';
  
  angular
    .module('harbor.project.member')
    .directive('switchRole', switchRole);
  
  SwitchRoleController.$inject = ['getRoleById'];
  
  function SwitchRoleController(getRoleById) {
    var vm = this;
    
    vm.currentRole = getRoleById(vm.roleId);            
    vm.selectRole = selectRole;
    
    function selectRole(role) {
      vm.currentRole = getRoleById(role.id);  
      vm.roleId = role.id;
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
        'roleId': '='
      },
      'controller' : SwitchRoleController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  } 
  
})();