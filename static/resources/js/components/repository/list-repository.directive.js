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
    
  ListRepositoryController.$inject = ['$scope', 'ListRepositoryService', 'DeleteRepositoryService', 'DeleteLabelService', '$filter', 'trFilter', '$location', 'getParameterByName'];
  
  function ListRepositoryController($scope, ListRepositoryService, DeleteRepositoryService, DeleteLabelService, $filter, trFilter, $location, getParameterByName) {
    
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
        
    vm.retrieve = retrieve;
    vm.tagCount = {};
    vm.labelCount = {};

    //初始化分页
    vm.paginationConf = {
      currentPage: 1,
      pagesLength: 5,
      itemsPerPage: 5,
      totalItems: 1,
      numberOfPages : 1,
      onChange: function(){
        //vm.refresh(vm.paginationConf.currentPage);
      }
   };
    
    vm.projectId = getParameterByName('project_id', $location.absUrl());
    vm.customId = getParameterByName('custom_id', $location.absUrl());
    vm.retrieve(); 
        
    $scope.$on('$locationChangeSuccess', function() {
      vm.projectId = getParameterByName('project_id', $location.absUrl());
      vm.customId = getParameterByName('custom_id', $location.absUrl());
      vm.filterInput = '';
      vm.retrieve();    
    });
    

    $scope.$watch('vm.repositories', function(current) {
      if(current) {
        vm.repositories = current || [];
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
                
    $scope.$on('labels', function(e, val) {
      vm.labels = val;
    });

    $scope.$on('label', function(e, val){
      vm.label = val;
    });

    $scope.$on('labelCount', function(e, val) {
      vm.labelCount = val;
    });

    $scope.$on('addedSuccess', function(e, val) {
      vm.retrieve();
    });

    $scope.$watch('vm.paginationConf.currentPage', function(e, val) {
      vm.retrieve();
    });

    vm.deleteByRepo = deleteByRepo;
    vm.deleteByTag = deleteByTag;
    vm.deleteByLabel = deleteByLabel;
    vm.deleteLabel = deleteLabel;
    vm.deleteImage =  deleteImage;
                
    function retrieve(){
      var pageId = vm.paginationConf.currentPage || 1;
      ListRepositoryService(vm.projectId, vm.filterInput, pageId, vm.customId)
        .success(getRepositoryComplete)
        .error(getRepositoryFailed);
    }
   
    //Depending on the configuration tab according to initialize
    function getRepositoryComplete(data, status) {
      vm.repositories = data.repoList || [];
      vm.paginationConf.totalItems = data.totalItems || 1;
      vm.paginationConf.numberOfPages =  data.pages || 1;
      vm.paginationConf.itemsPerPage = data.pagesize || 5;
      $scope.$broadcast('refreshTagsAndLabels', true);
    }
    
    function getRepositoryFailed(response) {
      console.log('Failed to list repositories:' + response);      
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

    function deleteByLabel() {
      $scope.$emit('modalTitle', $filter('tr')('alert_delete_label_title', [vm.label]));
      var message;
      $scope.$emit('modalMessage',  $filter('tr')('alert_delete_label', [vm.label]));

      var emitInfo = {
        'confirmOnly': false,
        'contentType': 'text/html',
        'action' : vm.deleteLabel
      };

      $scope.$emit('raiseInfo', emitInfo);
    }


    function deleteLabel() {
      console.log('Delete image, repoName:' + vm.repoName + ', label:' + vm.label);
      vm.toggleInProgress[vm.repoName + '|' + vm.label] = true;
      DeleteLabelService(vm.repoName, vm.label)
        .success(deleteLabelSuccess)
        .error(deleteLabelFailed);
    }

    function deleteLabelSuccess(data, status) {
      vm.toggleInProgress[vm.repoName + '|' + vm.label] = false;
      vm.retrieve();
    }

    function deleteLabelFailed(data, status) {
      vm.toggleInProgress[vm.repoName + '|' + vm.label] = false;

      $scope.$emit('modalTitle', $filter('tr')('error'));
      var message;
      if(status === 401) {
        message = $filter('tr')('failed_to_delete_repo_insuffient_permissions');
      }else{
        message = $filter('tr')('failed_to_delete_repo');
      }
      $scope.$emit('modalMessage', message);
      $scope.$emit('raiseError', true);
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