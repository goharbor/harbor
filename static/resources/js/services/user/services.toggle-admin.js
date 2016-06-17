(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.user')
    .factory('ToggleAdminService', ToggleAdminService);
    
  ToggleAdminService.$inject = ['$http'];
  
  function ToggleAdminService($http) {
    
    return toggleAdmin;
    
    function toggleAdmin(userId, enabled) {
      return $http
        .put('/api/users/' + userId + '/sysadmin', {
          'has_admin_role' : enabled
        });
    }
    
  }
  
})();