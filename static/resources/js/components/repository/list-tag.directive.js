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
    
    vm.deleteByTag = deleteByTag;
    
    function retrieve() {
      ListTagService(vm.repoName)
        .then(getTagComplete)
        .catch(getTagFailed);
    }
    
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
      $scope.$emit('modalTitle', $filter('tr')('alert_delete_tag_title', [e.tag]));
      
      var message;
      if(vm.tags.length === 1) {
        message = $filter('tr')('alert_delete_last_tag', [e.tag]);
      }else {
        message = $filter('tr')('alert_delete_tag', [e.tag]);
      }
      $scope.$emit('modalMessage', message);
    }
    
  }
  
  function listTag() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/repository/list-tag.directive.html',
      'scope': {
        'tagCount': '=',
        'associateId': '=',
        'repoName': '='
      },
      'replace': true,
      'link': link,
      'controller': ListTagController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      
    }
  }
  
})();