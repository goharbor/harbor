(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.user')
    .factory('ResetPasswordService', ResetPasswordService);
    
  ResetPasswordService.$inject = ['$http', '$log'];
    
  function ResetPasswordService($http, $log) {
    return resetPassword;
    function resetPassword(uuid, password) {
      return $http({
          method: 'POST',
          url: '/reset',
          headers: {'Content-Type': 'application/x-www-form-urlencoded'},
          transformRequest: function(obj) {
              var str = [];
              for(var p in obj)
              str.push(encodeURIComponent(p) + "=" + encodeURIComponent(obj[p]));
              return str.join("&");
          },
          data: {'reset_uuid': uuid, 'password': password}
      });
    } 
  }
  
})();