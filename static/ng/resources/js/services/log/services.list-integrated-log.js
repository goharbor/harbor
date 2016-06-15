(function() {
  'use strict';
 
   angular
    .module('harbor.services.log')
    .factory('ListIntegratedLogService', ListIntegratedLogService);
  
  ListIntegratedLogService.$inject = ['$http', '$log'];
  
  function ListIntegratedLogService($http, $log) {
    
    return listIntegratedLog;
    
    function listIntegratedLog(lines) {
      $log.info('Get recent logs of the projects which the user is a member of:');
      return $http
        .get('/api/logs', {
          'params' : {
            'lines': lines,
          }
        });
      
    }
  }
  
})();