(function() {
  'use strict';
  
  angular
    .module('harbor.services.project')
    .factory('AddProjectService', AddProjectService);
    
  AddProjectService.$inject = ['$http', '$log'];
    
  function AddProjectService($http, $log) {
    
    return AddProject;
    
    function AddProject(projectName, isPublic) {
        return $http
          .post('/api/projects', {
            'project_name': projectName,
            'public': isPublic
          });
    }
  }
  
})();