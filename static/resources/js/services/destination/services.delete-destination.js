(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.destination')
    .factory('DeleteDestinationService', DeleteDestinationService);
    
  DeleteDestinationService.$inject = ['$http'];
  
  function DeleteDestinationService($http) {
    return deleteDestination;
    function deleteDestination(targetId) {
      return $http
        .delete('/api/targets/' + targetId);
    }
  }
  
})();