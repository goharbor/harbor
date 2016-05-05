(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.repository')
    .factory('ListManifestService', ListManifestService);
    
  ListManifestService.$inject = ['$http', '$log'];
  
  function ListManifestService($http, $log) {
    return ListManifest;
    function ListManifest(repoName, tag) {
      return $http
        .get('/api/repositories/manifests', {
          'params': {
            'repo_name': repoName,
            'tag': tag
          }
        });
    }
  }
  
})();