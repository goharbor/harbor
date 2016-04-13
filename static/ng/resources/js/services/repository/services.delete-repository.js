(function() {
  'use strict';
  
  angular
    .module('harbor.services.repository')
    .factory('DeleteRepositoryService', DeleteRepositoryService);
    
  DeleteRepositoryService.$inject = ['$http', '$log'];
  
  function DeleteRepositoryService($http, $log) {
    
    return DeleteRepository;
    
    function DeleteRepository(repository) {
      
    }
    
  }
})();