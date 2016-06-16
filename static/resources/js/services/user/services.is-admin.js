(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.user')
    .factory('IsAdminService', IsAdminService);
    
  IsAdminService.$inject = ['$http', '$log'];
  
  function IsAdminService($http, $log) {
    
    return IsAdmin;
    
    function IsAdmin() {
      
    }
    
  }
  
})