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
    .module('harbor.top.repository')
    .directive('topRepository', topRepository);
    
  TopRepositoryController.$inject = ['$scope', 'ListTopRepositoryService', '$filter', 'trFilter'];
  
  function TopRepositoryController($scope, ListTopRepositoryService, $filter, trFilter) {
    var vm = this;
    
    ListTopRepositoryService(5)
      .success(listTopRepositorySuccess)
      .error(listTopRepositoryFailed);

    function listTopRepositorySuccess(data) {
      vm.top10Repositories = data || [];
    }

    function listTopRepositoryFailed(data, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_get_top_repo'));
      $scope.$emit('raiseError', true);
      console.log('Failed to get top repo:' + data);
    }
        
  }
  
  function topRepository() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/top-repository/top-repository.directive.html',
      'controller': TopRepositoryController,
      'scope' : {
        'customBodyHeight': '='
      },
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
    
})();
