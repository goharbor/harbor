(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .directive('switchRoles', switchRoles);
  
  
  SwitchRolesController.$inject = [];
  
  function SwitchRolesController() {
    var vm = this;
    
    vm.currentRole = getRoleById(vm.roleId);
    vm.selectRole = selectRole;
        
    function selectRole(role) {
      vm.roleId = role.id;
      vm.currentRole = getRoleById(vm.roleId);
    }
    
    function getRoleById(roleId) {
      for(var i = 0; i < vm.roles.length; i++) {
        var role = vm.roles[i];
        if(role.id == roleId) {
          return role;
        }
      }
    }
    
  }
  
  function switchRoles() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/details/switch-roles.directive.html',
      'scope': {
        'roles': '=',
        'editMode': '=',
        'userId': '=',
        'roleId': '='
      },
      'link' : link,
      'controller' : SwitchRolesController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
     
    }
  } 
  
})();