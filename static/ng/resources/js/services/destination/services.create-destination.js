(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.destination')
    .factory('CreateDestinationService', CreateDestinationService);
    
  CreateDestinationService.$inject = ['$http', '$q', '$timeout'];
  
  function CreateDestinationService($http, $q, $timeout) {
    return createDestination;
    function createDestination() {
      
    }
  }
  
})()