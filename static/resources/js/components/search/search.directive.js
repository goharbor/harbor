(function() {
  
  'use strict';
  
  angular
    .module('harbor.search')
    .directive('search', search);
    
  SearchController.$inject = ['SearchService', '$scope'];
  
  function SearchController(SearchService, $scope) {
    var vm = this;
    vm.keywords = "";
    vm.search = searchByFilter;
    vm.filterBy = 'repository';
    
    searchByFilter();
    
    
    function searchByFilter() {
      SearchService(vm.keywords)
        .success(searchSuccess)
        .error(searchFailed);
    }
    
    function searchSuccess(data, status) {
      console.log('filterBy:' + vm.filterBy + ", data:" + data);
      vm.searchResult = data[vm.filterBy];
    }
    
    function searchFailed(data, status) {
      console.log('Failed to search:' + data);
    }
    
  }
  
  function search() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/search/search.directive.html',
      'scope': {
        'filterBy': '='
      },
      'controller': SearchController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    
    return directive; 
  }
  
})();