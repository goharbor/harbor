(function() {
  'use strict';
  
  angular
    .module('harbor.repository')
    .directive('listRepository', listRepository);   
    
  ListRepositoryController.$inject = ['ListRepositoryService'];
  
  function ListRepositoryController(ListRepositoryService) {
  
  }
  
  function listRepository() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/ng/resources/js/components/repository/list-repository.directive.html',
      replace: true,
      scope: {
        info: '='
      },
      controller: ListRepositoryController,
      controllerAs: 'vm',
      bindToController: true
    }
    
    return directive;
  }
  
})();