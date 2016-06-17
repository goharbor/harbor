(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.project')
    .factory('GetProjectById', GetProjectById);
    
  GetProjectById.$inject = ['$http'];
  
  function GetProjectById($http) {
    
    return getProject;
    
    function getProject(id) {
      return $http
        .get('/api/projects/' + id);
    }
    
  }
  
})();