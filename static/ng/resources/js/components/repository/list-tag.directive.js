(function() {
  
  'use strict';
  
  angular
    .module('harbor.repository')
    .directive('listTag', listTag);
    
  ListTagController.$inject = ['ListTagService'];
  
  function ListTagController(ListTagService) {
   
  }
  
  function listTag() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/repository/list-tag.directive.html',
      'scope': {
        'associateId': '=',
        'repoName': '=',
        'tags': '='
      },
      'replace': true,
      'controller': ListTagController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();