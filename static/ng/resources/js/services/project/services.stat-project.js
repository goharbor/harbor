(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.project')
    .factory('StatProjectService', StatProjectService);
    
  StatProjectService.$inject = ['$http', '$log'];
  
  function StatProjectService($http, $log) {
   
    return StatProject;
    
    function StatProject() {
      $log.info('statistics projects and repositories');
      return $http
        .get('/api/statistics');
    }

  }
  
})();