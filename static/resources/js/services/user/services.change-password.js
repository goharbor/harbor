(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.user')
    .factory('ChangePasswordService', ChangePasswordService);
    
  ChangePasswordService.$inject = ['$http', '$log'];
  
  function ChangePasswordService($http, $log) {
    
    return ChangePassword;
    
    function ChangePassword(userId, oldPassword, newPassword) {
      return $http
        .put('/api/users/' + userId + '/password', {
          'old_password': oldPassword,
          'new_password': newPassword
        });
    }
    
  }
  
})();