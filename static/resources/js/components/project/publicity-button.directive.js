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
    .module('harbor.project')
    .directive('publicityButton', publicityButton);
  
  PublicityButtonController.$inject = ['$scope', 'ToggleProjectPublicityService', '$filter', 'trFilter'];
  
  function PublicityButtonController($scope, ToggleProjectPublicityService, $filter, trFilter) {
    var vm = this;
    vm.toggle = toggle;
    
    function toggle() {      
      vm.isPublic = vm.isPublic ? 0 : 1;
      ToggleProjectPublicityService(vm.projectId, vm.isPublic)
        .success(toggleProjectPublicitySuccess)
        .error(toggleProjectPublicityFailed);
    }
    
    function toggleProjectPublicitySuccess(data, status) {
      
      console.log('Successful toggle project publicity.');
    }
    
    function toggleProjectPublicityFailed(e, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      var message;
      if(status === 403) {
        message = $filter('tr')('failed_to_toggle_publicity_insuffient_permissions');
      }else{
        message = $filter('tr')('failed_to_toggle_publicity');
      }
      $scope.$emit('modalMessage', message);
      $scope.$emit('raiseError', true);
      
      vm.isPublic = vm.isPublic ? 0 : 1;
      console.log('Failed to toggle project publicity:' + e);
    }
  }

  function publicityButton() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/project/publicity-button.directive.html',
      'scope': {
        'isPublic': '=',
        'projectId': '='
      },
      'link': link,
      'controller': PublicityButtonController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attr, ctrl) {
      scope.$watch('vm.isPublic', function(current, origin) {
        if(current) {
          ctrl.isPublic = current;
        }
      });  
    }
  }
  
})();