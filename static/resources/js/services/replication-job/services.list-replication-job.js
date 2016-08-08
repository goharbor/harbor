/*
    Copyright (c) 2016 VMware, Inc. All Rights Reserved.
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
        
        http://www.apache.org/licenses/LICENSE-2.0
        
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/
(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.replication.job')
    .factory('ListReplicationJobService', ListReplicationJobService);
    
  ListReplicationJobService.$inject = ['$http'];
  
  function ListReplicationJobService($http) {
    
    return listReplicationJob;
    
    function listReplicationJob(policyId, repository, status, startTime, endTime, page, pageSize) {
      return $http
        .get('/api/jobs/replication/?page=' + page + '&page_size=' + pageSize, {
          'params': {
            'policy_id': policyId,
            'repository': repository,
            'status': status,
            'start_time': startTime,
            'end_time': endTime
          }
        });
    }
  }
})();