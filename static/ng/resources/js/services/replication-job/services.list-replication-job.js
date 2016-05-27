(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.replication.job')
    .factory('ListReplicationJobService', ListReplicationJobService);
    
  ListReplicationJobService.$inject = ['$http', '$q', '$timeout'];
  
  function ListReplicationJobService($http, $q, $timeout) {
    var mockData = [
      {
        'policy_id': 2,
        'job_id': 1,
        'name': 'Replicate Ubuntu:14.04',
        'operation': 'copy',
        'start_time': '2016-05-27 14:05:22',
        'status': 'failed'
      },
      {
        'policy_id': 1,
        'job_id': 2,
        'name': 'Replicate MySQL:5.6',
        'operation': 'copy',
        'start_time': '2016-05-27 15:15:22',
        'status': 'success'
      },
      {
        'policy_id': 1,
        'job_id': 3,
        'name': 'Replicate Alpine:1.1',
        'operation': 'copy',
        'start_time': '2016-05-27 13:15:22',
        'status': 'success'
      },
      {
        'policy_id': 2,
        'job_id': 4,
        'name': 'Replicate Alpine:1.1',
        'operation': 'copy',
        'start_time': '2016-05-27 13:15:22',
        'status': 'success'
      }
    ];
    return listReplicationJob;
    
    
    
    function listReplicationJob(policyId) {
      console.log('policyId=' + policyId);
      var defer = $q.defer();
      $timeout(function() {
        var result = [];
        for(var i = 0; i < mockData.length; i++) {
          var item = mockData[i];
          if(item['policy_id'] == policyId) {
            result.push(item);
          }
        }
        defer.resolve(result);
      });
      return defer.promise;
    }
  }
  
})();