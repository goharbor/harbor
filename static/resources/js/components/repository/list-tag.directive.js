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
    
    $scope.$watch('vm.repoName', function(current, origin) {    
      if(current) {
        console.log('vm.repoName in tags:' + current);
        vm.retrieve();
      }
    });
    
    $scope.$on('refreshTags', function(e, val) {
      if(val) {
        vm.retrieve();
      }
    });
    
    vm.deleteTag = deleteTag;
    
    function retrieve() {
      ListTagService(vm.repoName)
        .then(getTagComplete)
        .catch(getTagFailed);
    }
    
    function getTagComplete(response) {
      
      vm.tags = response.data;
      vm.tagCount[vm.repoName] = vm.tags.length;
      
      $scope.$emit('tags', vm.tags);
      $scope.$emit('tagCount', vm.tagCount);
      
      angular.forEach(vm.tags, function(item) {
        vm.toggleInProgress[vm.repoName + '|' + item] = false;
      });
    }
      
    function getTagFailed(response) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_get_tag') + response);
      $scope.$emit('raiseError', true);
      console.log('Failed to get tag:' + response);
    }
    
    function deleteTag(e) {
      $scope.$emit('repoName', e.repoName); 
      $scope.$emit('tag', e.tag);
      vm.deleteByTag();
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
        'deleteByTag': '&'
      },
      'replace': true,
      'controller': ListTagController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;

  }
  
})();