(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.destination')
    .factory('ListDestinationPolicyService', ListDestinationPolicyService);
    
  ListDestinationPolicyService.$inject = ['$http'];
  
  function ListDestinationPolicyService($http) {
    return listDestinationPolicy;
    function listDestinationPolicy(targetId) {
      return $http
        .get('/api/targets/' + targetId + '/policies/');
    }
  }
  
})();