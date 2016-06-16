(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.destination')
    .factory('ListDestinationService', ListDestinationService);
    
  ListDestinationService.$inject = ['$http'];
  
  function ListDestinationService($http) {    
    return listDestination;
    function listDestination(targetId, name) {
      return $http
        .get('/api/targets/' + targetId, {
          'params': {
            'name': name
          }
        });
    }
  }
  
})()