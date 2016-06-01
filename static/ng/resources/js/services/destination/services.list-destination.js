(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.destination')
    .factory('ListDestinationService', ListDestinationService);
    
  ListDestinationService.$inject = ['$http', '$q', '$timeout'];
  
  function ListDestinationService($http, $q, $timeout) {
    
    var mockData = [
      {
        'id' : 1,
        'name': 'Target01',
        'endpoint': 'http://10.117.170.69',
        'creation_time': '2016-06-01 16:54:32'
      },
      {
        'id' : 2,
        'name': 'Target02',
        'endpoint': 'http://10.117.171.41',
        'creation_time': '2016-06-01 15:35:22'
      },
      {
        'id' : 3,
        'name': 'Target03',
        'endpoint': 'http://10.117.171.63',
        'creation_time': '2016-06-01 14:22:21'
      }
    ];
    
    return listDestination;
    function listDestination() {
      var q = $q.defer();
      $timeout(function() {
        q.resolve(mockData);
      });      
      return q.promise;
    }
  }
  
})()