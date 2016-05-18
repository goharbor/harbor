(function() {
  
  'use strict';
  
  angular
    .module('harbor.project.member')
    .directive('switchRole', switchRole);
  
  SwitchRoleController.$inject = ['getRole', '$scope'];
  
  function SwitchRoleController(getRole, $scope) {
    var vm = this;
    
    $scope.$watch('vm.roleName', function(current,origin) {
      if(current) {
        vm.currentRole = getRole({'key': 'roleName', 'value': current});            
      }
    });
    vm.selectRole = selectRole;
    
    function selectRole(role) {
      vm.currentRole = getRole({'key': 'roleName', 'value': role.roleName});  
      vm.roleName = role.roleName;
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