(function() {
  'use strict';
 
   angular
    .module('harbor.services.project')
    .factory('ListProjectService', ListProjectService);
  
  ListProjectService.$inject = ['$http', '$log'];
  
  function ListProjectService($http, $log) {
    
    return ListProject;
    
    function ListProject(queryParams) {
      
      $log.info(queryParams);
      
      var isPublic = queryParams.isPublic;      
      var projectName = queryParams.projectName;
      
      return $http
        .get('/api/projects',{
          params: {
            'is_public': isPublic,
            'project_name': projectName
          }
        });
      
    }
  }
})();