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
    .directive('listTag', listTag);
    
  ListTagController.$inject = ['$scope', 'ListTagService', '$filter', 'trFilter'];
  
  function ListTagController($scope, ListTagService, $filter, trFilter) {
    var vm = this;
    
    vm.tags = [];
    vm.retrieve = retrieve;
    
    vm.selected = []
    vm.selected[vm.repoName] = [];
    
    vm.selectedTags = [];
    
    $scope.$watch('vm.repoName', function(current, origin) {    
      if(current) {
        console.log('vm.repoName triggered tags retrieval.')
        vm.retrieve();
      }
    });
    
    $scope.$on('refreshTags', function(e, val) {
      if(val) {
        vm.retrieve();
        vm.selectedCount[vm.repoName] = 0;
        vm.selected[val.repoName] = []; 
        vm.selectedTags = [];
      }
    });
    
    $scope.$watch('vm.selectedCount[vm.repoName]', function(current, previous) {
      if(current !== previous) {
        console.log('Watching vm.selectedCount:' + current);
        $scope.$emit('selectedAll', {'status': (current === vm.tags.length), 'repoName': vm.repoName});
      }
    });
    
    $scope.$on('gatherSelectedTags' + vm.repoName, function(e, val) {
      if(val) {
        console.log('RECEIVED gatherSelectedTags:' + val);
        gatherSelectedTags();
      }
    })
                    
    $scope.$on('selectAll' + vm.repoName, function(e, val) {      
      (val.status) ? vm.selectedCount[val.repoName] = vm.tags.length : vm.selectedCount[val.repoName] = 0;
      for(var i = 0; i < vm.tags.length; i++) {
        vm.selected[val.repoName][i] = val.status;
      }
      gatherSelectedTags();
      console.log('received selectAll:' + angular.toJson(val) + ', vm.selected:' + angular.toJson(vm.selected));
    });              
     
    $scope.$watch('vm.tags', function(current) {
      if(current) {
        vm.tags = current;
      }
    });
    
    vm.deleteTag = deleteTag;    
    
    vm.selectedCount = [];
    vm.selectedCount[vm.repoName] = 0;
    
    vm.selectOne = selectOne;   
    
    function retrieve() {
      ListTagService(vm.repoName)
        .success(getTagSuccess)
        .error(getTagFailed);
    }
    
    function getTagSuccess(data) {
      vm.tags = data || [];
      vm.tagCount[vm.repoName] = vm.tags.length;
      
      $scope.$emit('tags', vm.tags);
      $scope.$emit('tagCount', vm.tagCount);
      
      angular.forEach(vm.tags, function(item) {
        vm.toggleInProgress[vm.repoName + '|' + item] = false;
      });
      
      for(var i = 0; i < vm.tags.length; i++) {
        vm.selected[vm.repoName][i] = false;
      }
    }
      
    function getTagFailed(data) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_get_tag'));
      $scope.$emit('raiseError', true);
      console.log('Failed to get tags:' + data);
    }
    
    function deleteTag(e) {
      $scope.$emit('repoName', e.repoName); 
      $scope.$emit('tag', e.tag);
      vm.deleteByTag();
    }    
    
    function selectOne(index, tagName) {      
      vm.selected[vm.repoName][index] = !vm.selected[vm.repoName][index];
      (vm.selected[vm.repoName][index]) ? ++vm.selectedCount[vm.repoName] : --vm.selectedCount[vm.repoName];
      console.log('selectOne, repoName:' + vm.repoName + ', vm.selected:' + vm.selected[vm.repoName][index]  + ', index:' + index + ', length:' + vm.selectedCount[vm.repoName]);        
      gatherSelectedTags();
    }
    
    function gatherSelectedTags() {
      vm.selectedTags[vm.repoName] = [];
      for(var i = 0; i < vm.tags.length; i++) {
        (vm.selected[vm.repoName][i]) ? vm.selectedTags[vm.repoName][i] = vm.tags[i] : vm.selectedTags[vm.repoName][i] = '';
      }
      var tagsToDelete = [];
      for(var i in vm.selectedTags[vm.repoName]) {
        var tag = vm.selectedTags[vm.repoName][i];
        if(tag !== '') {
          tagsToDelete.push(tag);
        }
      }
      $scope.$emit('selectedTags', {'repoName': vm.repoName, 'tags': tagsToDelete}); 
    }
    
  }
  
  function listTag() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/repository/list-tag.directive.html',
      'scope': {
        'tagCount': '=',
        'associateId': '=',
        'repoName': '=',
        'toggleInProgress': '=',
        'deleteByTag': '&',
        'roleId': '@'
      },
      'replace': true,
      'controller': ListTagController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;

  }
  
})();