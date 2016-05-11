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
    
    vm.deleteTag = deleteTag;
    
    function getTagComplete(response) {
      vm.tags = response.data;
    }
      
    function getTagFailed(response) {
      console.log('Failed get tag:' + response);
    }
    
    function deleteTag(e) {
      $scope.$emit('tag', e.tag);
      $scope.$emit('repoName', e.repoName);
    }
    
  }
  
  function listTag() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/repository/list-tag.directive.html',
      'scope': {
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