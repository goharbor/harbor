(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.repository')
    .factory('ListTagService', ListTagService);
  
  ListTagService.$inject = ['$http', '$log'];
  
  function ListTagService($http, $log) {
    return ListTag;
    
    function ListTag(repoName) {
      return $http
        .get('/api/repositories/tags', {
          'params': {
            'repo_name': repoName            
          }
        });
    }
  }
  
  
})();