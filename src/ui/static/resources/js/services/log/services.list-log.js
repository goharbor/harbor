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
    .module('harbor.services.log')
    .factory('ListLogService', ListLogService);
  
  ListLogService.$inject = ['$http', '$log'];
  
  function ListLogService($http, $log) {
    
    return LogResult;
    
    function LogResult(queryParams, page, pageSize) {      
      var projectId = queryParams.projectId;
      var username = queryParams.username;
      var beginTimestamp = queryParams.beginTimestamp;
      var endTimestamp = queryParams.endTimestamp;
      var keywords = queryParams.keywords;
      
      return $http
        .post('/api/projects/' + projectId + '/logs/filter?page=' + page + '&page_size=' + pageSize, {
          'begin_timestamp' : beginTimestamp,
          'end_timestamp'   : endTimestamp,
          'keywords' : keywords,
          'project_id': Number(projectId),
          'username' : username
        });
    }
  }
})();