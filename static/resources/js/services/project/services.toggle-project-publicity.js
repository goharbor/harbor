(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.project')
    .factory('ToggleProjectPublicityService', ToggleProjectPublicityService);
    
  ToggleProjectPublicityService.$inject = ['$http'];
  
  function ToggleProjectPublicityService($http) {
    return toggleProjectPublicity;
    function toggleProjectPublicity(projectId, isPublic) {
      return $http
        .put('/api/projects/' + projectId + '/publicity', {
          'public': isPublic
        });
    }
    
  }
  
})();