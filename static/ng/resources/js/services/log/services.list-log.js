(function() {
  
  'use strict';
 
   angular
    .module('harbor.services.log')
    .factory('ListLogService', ListLogService);
  
  ListLogService.$inject = ['$http', '$log'];
  
  function ListLogService($http, $log) {
    
    return LogResult;
    
    function LogResult(queryParams) {      
      $log.info(queryParams);
    }
  }
})();