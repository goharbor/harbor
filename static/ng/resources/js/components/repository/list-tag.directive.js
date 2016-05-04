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
    function getTagComplete(response) {
      vm.tags = response.data;
    }
      
    function getTagFailed(response) {
      
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