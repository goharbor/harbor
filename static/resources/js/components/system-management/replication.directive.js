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
    .module('harbor.system.management')
    .directive('replication', replication);
  
  ReplicationController.$inject = ['$scope', 'ListReplicationPolicyService', 'ToggleReplicationPolicyService', '$filter', 'trFilter'];
  
  function ReplicationController($scope, ListReplicationPolicyService, ToggleReplicationPolicyService, $filter, trFilter) {
    
    $scope.subsSubPane = 276;
    
    var vm = this;
    vm.retrieve = retrieve;
    vm.search = search;
    vm.confirmToTogglePolicy = confirmToTogglePolicy;
    vm.togglePolicy = togglePolicy;
    vm.editReplication = editReplication;
    vm.retrieve();
    
    function search() {
      vm.retrieve();
    }
    
    function retrieve() {
      ListReplicationPolicyService('', '', vm.replicationName)
        .success(listReplicationPolicySuccess)
        .error(listReplicationPolicyFailed);
    }
    
    function listReplicationPolicySuccess(data, status) {
      vm.replications = data || [];
    }
    
    function listReplicationPolicyFailed(data, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_list_replication'));
      $scope.$emit('raiseError', true);
      console.log('Failed to list replication policy.');
    }

    function confirmToTogglePolicy(policyId, enabled, name) {
      vm.policyId = policyId;
      vm.enabled = enabled;

      var status = $filter('tr')(vm.enabled === 1 ? 'enable':'disable');

      var title;
      var message;
      if(enabled === 1){
        title = $filter('tr')('confirm_to_toggle_enabled_policy_title');
        message = $filter('tr')('confirm_to_toggle_enabled_policy');
      }else{
        title = $filter('tr')('confirm_to_toggle_disabled_policy_title');
        message = $filter('tr')('confirm_to_toggle_disabled_policy');
      }
      $scope.$emit('modalTitle', title);
      $scope.$emit('modalMessage', message);
            
      var emitInfo = {
        'contentType': 'text/html',
        'confirmOnly': false,
        'action': vm.togglePolicy
      };
      
      $scope.$emit('raiseInfo', emitInfo);
    }
    
    function togglePolicy() {
      ToggleReplicationPolicyService(vm.policyId, vm.enabled)
        .success(toggleReplicationPolicySuccess)
        .error(toggleReplicationPolicyFailed);
    }
    
    function toggleReplicationPolicySuccess(data, status) {
      console.log('Successful toggle replication policy.');
      vm.retrieve();
    }
    
    function toggleReplicationPolicyFailed(data, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_toggle_policy'));
      $scope.$emit('raiseError', true);
      console.log('Failed to toggle replication policy.');
    }
    
    function editReplication(policyId) {
      vm.action = 'EDIT';
      vm.policyId = policyId;
    }
  }
  
  function replication() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/system-management/replication.directive.html',
      'scope': true,
      'link': link,
      'controller': ReplicationController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      element.find('#txtSearchInput').on('keydown', function(e) {
        if($(this).is(':focus') && e.keyCode === 13) {
          ctrl.retrieve();
        }
      });
    }
  }
  
})();
