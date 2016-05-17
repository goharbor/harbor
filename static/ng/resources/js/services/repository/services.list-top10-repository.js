(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.repository')
    .factory('ListTop10RepositoryService', ListTop10RepositoryService);
    
  ListTop10RepositoryService.$inject = ['$http', '$q', '$timeout'];
    
  function ListTop10RepositoryService($http, $q, $timeout) {
    
    var mockData = [
      {
        'repo_name': 'myproject/ubuntu',
        'image_size': '89',
        'creator': 'kunw'
      },
      {
        'repo_name': 'myproject/nginx',
        'image_size': '67',
        'creator': 'kunw'
      },
      {
        'repo_name': 'myrepo/mysql',
        'image_size': '122',
        'creator': 'user1'
      },
      {
        'repo_name': 'target/golang',
        'image_size': '587',
        'creator': 'tester'
      }
    ];
   
    function async() {
      
      var deferred = $q.defer();   
      
      $timeout(function() {
        deferred.resolve(mockData);
      }, 500);
      
      return deferred.promise;
    }
    
    return listTop10Repository;
    
    function listTop10Repository() {
      return async();
    }
    
  }
  
  
  
  
})();