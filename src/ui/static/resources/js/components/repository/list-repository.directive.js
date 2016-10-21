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
    .module('harbor.repository')
    .directive('listRepository', listRepository);   
    
  ListRepositoryController.$inject = ['$scope', 'ListRepositoryService', 'DeleteRepositoryService', '$filter', 'trFilter', '$location', 'getParameterByName'];
  
  function ListRepositoryController($scope, ListRepositoryService, DeleteRepositoryService, $filter, trFilter, $location, getParameterByName) {
    
    $scope.subsTabPane = 30;
    
    var vm = this;
  
    vm.sectionHeight = {'min-height': '579px'};
  
    vm.filterInput = '';
    vm.toggleInProgress = [];
       
    var hashValue = $location.hash();
    if(hashValue) {
      var slashIndex = hashValue.indexOf('/');
      if(slashIndex >=0) {
        vm.filterInput = hashValue.substring(slashIndex + 1);      
      }else{
        vm.filterInput = hashValue;
      }
    }
    vm.page = 1;
    vm.pageSize = 15;    
    
    vm.retrieve = retrieve;
    vm.searchRepo = searchRepo;
    vm.tagCount = {};
    
    vm.projectId = getParameterByName('project_id', $location.absUrl());
        
    $scope.$on('retrieveData', function(e, val) {
      if(val) {
        vm.projectId = getParameterByName('project_id', $location.absUrl());
        vm.filterInput = '';
        vm.retrieve();        
      }
    });
     

    $scope.$watch('vm.repositories', function(current) {
      if(current) {
        vm.repositories = current || [];
      }
    });
        
    $scope.$watch('vm.page', function(current) {
      if(current) {
        vm.page = current;
        vm.retrieve();
      }
    });
    
    $scope.$on('repoName', function(e, val) {
      vm.repoName = val;
    });

    $scope.$on('tag', function(e, val){
      vm.tag = val;
    });
    
    $scope.$on('tagCount', function(e, val) {
      vm.tagCount = val;
    });
        
    $scope.$on('tags', function(e, val) {
      vm.tags = val;
    });
            
    vm.deleteByRepo = deleteByRepo;
    vm.deleteByTag = deleteByTag;
    vm.deleteImage =  deleteImage;
                
    function retrieve(){
      console.log('retrieve repositories, project_id:' + vm.projectId);
      ListRepositoryService(vm.projectId, vm.filterInput, vm.page, vm.pageSize)
        .then(getRepositoryComplete, getRepositoryFailed);
    }
   
    function getRepositoryComplete(response) {
      vm.repositories = response.data || [];
      vm.totalCount = response.headers('X-Total-Count');
    }
    
    function getRepositoryFailed(response) {
      console.log('Failed to list repositories:' + response);      
    }
   
    function searchRepo() {
      $scope.$broadcast('refreshTags', true);
      vm.retrieve();
    }
  
    function deleteByRepo(repoName) { 
      vm.repoName = repoName;
      vm.tag = '';
      
      $scope.$emit('modalTitle', $filter('tr')('alert_delete_repo_title', [repoName]));
      $scope.$emit('modalMessage', $filter('tr')('alert_delete_repo', [repoName]));
      
      var emitInfo = {
        'confirmOnly': false,
        'contentType': 'text/html',
        'action' : vm.deleteImage
      };
      
      $scope.$emit('raiseInfo', emitInfo);
    }
    
    function deleteByTag() {
      $scope.$emit('modalTitle', $filter('tr')('alert_delete_tag_title', [vm.tag]));
      var message;
      console.log('vm.tagCount:' + angular.toJson(vm.tagCount[vm.repoName]));
      $scope.$emit('modalMessage',  $filter('tr')('alert_delete_tag', [vm.tag]));
      
      var emitInfo = {
        'confirmOnly': false,
        'contentType': 'text/html',
        'action' : vm.deleteImage
      };
      
      $scope.$emit('raiseInfo', emitInfo);
    }
  
    function deleteImage() {
      
      console.log('Delete image, repoName:' + vm.repoName + ', tag:' + vm.tag);
      vm.toggleInProgress[vm.repoName + '|' + vm.tag] = true;
      DeleteRepositoryService(vm.repoName, vm.tag)
        .success(deleteRepositorySuccess)
        .error(deleteRepositoryFailed);
    }
    
    function deleteRepositorySuccess(data, status) {
      vm.toggleInProgress[vm.repoName + '|' + vm.tag] = false;
      vm.retrieve();
    }
    
    function deleteRepositoryFailed(data, status) {
      vm.toggleInProgress[vm.repoName + '|' + vm.tag] = false;  
        
      $scope.$emit('modalTitle', $filter('tr')('error'));
      var message;
      if(status === 401) {
        message = $filter('tr')('failed_to_delete_repo_insuffient_permissions');
      }else{
        message = $filter('tr')('failed_to_delete_repo');
      }
      $scope.$emit('modalMessage', message);
      $scope.$emit('raiseError', true);
      
      console.log('Failed to delete repository:' + angular.toJson(data));
    }
    
  }
  
  function listRepository() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/repository/list-repository.directive.html',
      'scope': {
        'sectionHeight': '='
      },
      'link': link,
      'controller': ListRepositoryController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    
    return directive;
    
    function link(scope, element, attr, ctrl) {
      element.find('#txtSearchInput').on('keydown', function(e) {
        if($(this).is(':focus') && e.keyCode === 13) {
          ctrl.retrieve();
        }
      });
    }
  
  }
  
})();