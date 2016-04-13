(function() {
  'use strict';
 
   angular
    .module('harbor.services.project')
    .factory('ListProjectService', ListProjectService);
  
  ListProjectService.$inject = ['$http', '$log'];
  
  function ListProjectService($http, $log) {
    
    return ListProject;
    
    function ListProject(queryParams) {      
      $log.info(queryParams);
    }
  }
})();