(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.search')
    .factory('SearchService', SearchService);
    
  SearchService.$inject = ['$http', '$log'];
  
  function SearchService($http, $log) {
    
    return Search;
    
    function Search(queryParams) {
      
    }
    
  }
  
})();