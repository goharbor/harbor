(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.replication.policy')
    .factory('ListReplicationPolicyService', ListReplicationPolicyService);
    
  ListReplicationPolicyService.$inject = ['$http', '$q', '$timeout'];
  
  function ListReplicationPolicyService($http, $q, $timeout) {
    
    var mockData = [
      {
        'name': 'test01',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-26 22:30:30',
        'status': 'NORMAL',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      },
      {
        'name': 'test02',
        'description': 'Sync image for current.',
        'destination': '10.117.170.69',
        'start_time': '2015-05-27 20:00:00',
        'status': 'WARNING',
        'activation': true
      }
      
    ];
    
    return listReplicationPolicy;
    
    function async() {
      var defer = $q.defer();
      $timeout(function() {
        defer.resolve(mockData);
      });
      return defer.promise;
    }
    
    function listReplicationPolicy(policyName) {
      return async();
    }
    
  }
  
})();