/*
    Copyright (c) 2016 VMware, Inc. All Rights Reserved.
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
        
        http://www.apache.org/licenses/LICENSE-2.0
        
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/
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
      'templateUrl': '/static/resources/js/components/project-member/switch-role.directive.html',
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