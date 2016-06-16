(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.search')
    .controller('SearchController', SearchController);
   
  SearchController.$inject = ['$window', 'SearchService'];
  
  function SearchController($window, SearchService) {
    var vm = this;
    if($window.location.search) {
      vm.q = $window.location.search.split('=')[1];
      console.log('vm.q:' + vm.q);
      SearchService(vm.q)
        .success(searchSuccess)
        .error(searchFailed);
    }
    
    function searchSuccess(data, status) {
      vm.repository = data['repository'];
      vm.project = data['project'];
    }
    
    function searchFailed(data, status) {
      console.log('Failed search:' + data);
    }
  }
  
})();