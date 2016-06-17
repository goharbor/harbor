(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.destination')
    .factory('UpdateDestinationService', UpdateDestinationService);
    
  UpdateDestinationService.$inject = ['$http'];  
    
  function UpdateDestinationService($http) {
    return updateDestination;
    function updateDestination(targetId, target) {
      return $http
        .put('/api/targets/' + targetId, {
          'name': target.name,
          'endpoint': target.endpoint,
          'username': target.username,
          'password': target.password
        });
    }
  }
  
})();