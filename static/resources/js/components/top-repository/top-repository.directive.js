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
    
  TopRepositoryController.$inject = ['$scope', 'ListTopRepositoryService', 'SearchService', '$filter', 'trFilter', '$window'];
  
  function TopRepositoryController($scope, ListTopRepositoryService, SearchService, $filter, trFilter, $window) {
    var vm = this;
    
    ListTopRepositoryService(5)
      .success(listTopRepositorySuccess)
      .error(listTopRepositoryFailed);

    vm.gotoRepo = gotoRepo;
      
    function listTopRepositorySuccess(data) {
      vm.top10Repositories = data || [];
    }

    function listTopRepositoryFailed(data, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_get_top_repo'));
      $scope.$emit('raiseError', true);
      console.log('Failed to get top repo:' + data);
    }
    
    function gotoRepo(repoName) {
      SearchService(repoName)
        .success(searchSuccess)
        .error(searchFailed);
    }
    
    function searchSuccess(data, status) {
      var repoInfo = data['repository'];
      if(repoInfo && repoInfo.length > 0) {
        var projectId = repoInfo[0]['project_id'];
        var publicity = repoInfo[0]['project_public'];
        var repoName = repoInfo[0]['repository_name'];
        $window.location.href = '/repository#/repositories?project_id=' + projectId + '&is_public=' + publicity +'#' + repoName;
      }
    }
    
    function searchFailed(data) {
      console.log('Failed to get repo info:' + data);
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
