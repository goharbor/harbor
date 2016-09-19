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
    .module('harbor.user.log')
    .directive('userLog', userLog);
    
  UserLogController.$inject = ['$scope', 'ListIntegratedLogService', '$filter', 'trFilter', '$window'];
  
  function UserLogController($scope, ListIntegratedLogService, $filter, trFilter, $window) {
    var vm = this;
    
    ListIntegratedLogService()
      .success(listIntegratedLogSuccess)
      .error(listIntegratedLogFailed);
    
    vm.gotoLog = gotoLog;
    
    function listIntegratedLogSuccess(data) {
      vm.integratedLogs = data || [];
    }

    function listIntegratedLogFailed(data, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_get_user_log') + data);
      $scope.$emit('raiseError', true);
      console.log('Failed to get user logs:' + data);
    }
    
    function gotoLog(projectId, username) {
      $window.location.href = '/repository#/logs?project_id=' + projectId + '#' + encodeURIComponent(username);
    }
    
  }
  
  function userLog() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/user-log/user-log.directive.html',
      'controller': UserLogController,
      'scope' : true,
      'controllerAs': 'vm',
      'bindToController': true
    };
    
    return directive;
  }
    
})();
