(function() {
  'use strict';
 
   angular
    .module('harbor.services.project')
    .factory('ListProjectService', ListProjectService);
  
  ListProjectService.$inject = ['$http', '$log'];
  
  function ListProjectService($http, $log) {
    
    return ListProject;
    
    function ListProject(projectName, isPublic) {
      $log.info('list project projectName:' + projectName, ', isPublic:' + isPublic);
      return $http
        .get('/api/projects', {
          'params' : {
            'is_public': isPublic,
            'project_name': projectName
          }
        });
      
    }
  }
})();