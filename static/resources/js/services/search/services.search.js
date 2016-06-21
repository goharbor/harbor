(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.search')
    .factory('SearchService', SearchService);
    
  SearchService.$inject = ['$http', '$log'];
  
  function SearchService($http, $log) {
    
    return search;
    
    function search(keywords) {
      return $http
        .get('/api/search',{
          params: {
            'q': keywords
          }
        });
    }
    
  }
  
})();