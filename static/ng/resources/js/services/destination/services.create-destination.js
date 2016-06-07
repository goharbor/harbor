(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.destination')
    .factory('CreateDestinationService', CreateDestinationService);
    
  CreateDestinationService.$inject = ['$http'];
  
  function CreateDestinationService($http) {
    return createDestination;
    function createDestination(name, endpoint, username, password) {
      return $http
        .post('/api/targets', {
          'name': name,
          'url': endpoint,
          'username': username,
          'password': password
        });
    }
  }
  
})()