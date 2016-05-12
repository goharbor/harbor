(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.project')
    .factory('StatProjectService', StatProjectService);
    
  StatProjectService.$inject = ['$http', '$q', '$timeout'];
  
  function StatProjectService($http, $q, $timeout) {
    
    var mockData = {
      'projects': 30,
      'public_projects': 50,
      'total_projects': 120,
      'repositories': 50,
      'public_repositories': 40,
      'total_repositories': 110
    };
    
    function async() {
      var deferred = $q.defer();
      
      $timeout(function() {
        deferred.resolve(mockData);
      }, 500);
      
      return deferred.promise;
    }
    
    return statProject;
    
    function statProject() {
      return async();
    }
    
  }
  
})();