(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.search')
    .factory('SearchService', SearchService);
    
  SearchService.$inject = ['$http', '$log'];
  
  function SearchService($http, $log) {
    
    return Search;
    
    function Search(keywords) {
      return $http({
          method: 'GET',
          url: '/api/search',
       
          transformRequest: function(obj) {
              var str = [];
              for(var p in obj)
              str.push(encodeURIComponent(p) + "=" + encodeURIComponent(obj[p]));
              return str.join("&");
          },
          data: {'q': keywords}
      });
    }
    
  }
  
})();