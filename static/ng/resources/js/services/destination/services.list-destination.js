(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.destination')
    .factory('ListDestinationService', ListDestinationService);
    
  ListDestinationService.$inject = ['$http'];
  
  function ListDestinationService($http) {    
    return listDestination;
    function listDestination(name) {
      return $http
        .get('/api/targets', {
          'params': {
            'name': name
          }
        });
    }
  }
  
})()