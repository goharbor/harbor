(function() {
  'use strict';
  
  angular
    .module('harbor.services.project')
    .factory('EditProjectService', EditProjectService);
    
  EditProjectService.$inject = ['$http', '$log'];
    
  function EditProjectService($http, $log) {
    
    return EditProject;
    
    function EditProject(projectId, isPublic) {
        return $http
          .put('/api/projects/' + projectId, {
            'public': isPublic
          });
    }
  }
  
})();