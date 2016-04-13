(function() {
  
  'use strict';
  
  angular
    .module('harbor.search')
    .controller('SearchController', SearchController);
    
  SearchController.$inject = ['SearchService'];
  
  function SearchController(SearchService) {
       
  }
  
})();