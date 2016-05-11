(function() {
  
  'use strict';
  
  angular
    .module('harbor.repository')
    .directive('listTag', listTag);
    
  ListTagController.$inject = ['$scope', 'ListTagService'];
  
  function ListTagController($scope, ListTagService) {
    var vm = this;
    
    vm.tags = [];
    
    $scope.$watch('vm.repoName', function(current, origin) {    
      if(current) {
        console.log('vm.repoName in tags:' + current);
        ListTagService(current)
          .then(getTagComplete)
          .catch(getTagFailed);
      }
    });
    
    vm.deleteByTag = deleteByTag;
    
    function getTagComplete(response) {
      vm.tags = response.data;
      vm.tagCount[vm.repoName] = vm.tags.length;
      $scope.$emit('tagCount', vm.tagCount);
    }
      
    function getTagFailed(response) {
      console.log('Failed get tag:' + response);
    }
    
    function deleteByTag(e) {
      $scope.$emit('tag', e.tag);
      $scope.$emit('repoName', e.repoName);
      $scope.$emit('modalTitle', 'Delete tag - ' + e.tag);
      
      var message;
      if(vm.tags.length == 1) {
        message = 'After deleting the associated repository with the tag will be deleted together,<br/>' +
        'because a repository contains at least one tag. And the corresponding image will be removed from the system.<br/>' +
        '<br/>Delete this "' + e.tag + '" tag now?';
      }else {
        message = 'Delete this "' + e.tag + '" tag now?';
      }
      
      $scope.$emit('modalMessage', message);
    }
    
  }
  
  function listTag() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/repository/list-tag.directive.html',
      'scope': {
        'tagCount': '=',
        'associateId': '=',
        'repoName': '='
      },
      'replace': true,
      'controller': ListTagController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();