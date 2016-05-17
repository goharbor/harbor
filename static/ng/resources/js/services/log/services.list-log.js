(function() {
  
  'use strict';
 
   angular
    .module('harbor.services.log')
    .factory('ListLogService', ListLogService);
  
  ListLogService.$inject = ['$http', '$log'];
  
  function ListLogService($http, $log) {
    
    return LogResult;
    
    function LogResult(queryParams) {      
      var projectId = queryParams.projectId;
      var username = queryParams.username;
      var beginTimestamp = queryParams.beginTimestamp;
      var endTimestamp = queryParams.endTimestamp;
      var keywords = queryParams.keywords;
      
      return $http
        .post('/api/projects/' + projectId + '/logs/filter', {
          'beginTimestamp' : beginTimestamp,
          'endTimestamp'   : endTimestamp,
          'keywords' : keywords,
          'project_id': projectId,
          'username' : username
        });
    }
  }
})();