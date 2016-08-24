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
    .module('harbor.user')
    .directive('toggleAdmin', toggleAdmin);
   
  ToggleAdminController.$inject = ['$scope', 'ToggleAdminService', '$filter', 'trFilter'];
  
  function ToggleAdminController($scope, ToggleAdminService, $filter, trFilter) {
    var vm = this;
    
    vm.isAdmin = (vm.hasAdminRole === 1);
    vm.enabled = vm.isAdmin ? 0 : 1;
    vm.toggle = toggle;
    vm.editable = (vm.currentUser.user_id !== Number(vm.userId));
    
    function toggle() {
      ToggleAdminService(vm.userId, vm.enabled)
        .success(toggleAdminSuccess)
        .error(toggleAdminFailed);        
    }    
    
    function toggleAdminSuccess(data, status) {
      if(vm.isAdmin) {
        vm.isAdmin = false;
      }else{
        vm.isAdmin = true;
      }
      console.log('Toggled userId:' + vm.userId + ' to admin:' + vm.isAdmin);
    }

    function toggleAdminFailed(data, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_toggle_admin'));
      $scope.$emit('raiseError', true);
      if(vm.isAdmin) {
        vm.isAdmin = false;
      }else{
        vm.isAdmin = true;
      }
      console.log('Failed to toggle admin:' + data);
    }    
  }
  
  function toggleAdmin() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/user/toggle-admin.directive.html',
      'scope': {
        'hasAdminRole': '=',
        'userId': '@',
        'currentUser': '='
      },
      'link': link,
      'controller': ToggleAdminController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      
    }
  }
  
})();
