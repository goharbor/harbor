(function() {
  'use strict';
  
  angular
    .module('harbor.services.repository')
    .factory('DeleteRepositoryService', DeleteRepositoryService);
    
  DeleteRepositoryService.$inject = ['$http', '$log'];
  
  function DeleteRepositoryService($http, $log) {
    
    return DeleteRepository;
    
    function DeleteRepository(repoName, tag) {
      var params = (tag === '') ? {'repo_name' : repoName} : {'repo_name': repoName, 'tag': tag};
      return $http
        .delete('/api/repositories', {
          'params': params
        });
    }
  }
})();