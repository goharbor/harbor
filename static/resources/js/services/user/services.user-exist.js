(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.user')
    .factory('UserExistService', UserExistService);
    
  UserExistService.$inject = ['$http', '$log'];
   
  function UserExistService($http, $log) {
    return userExist;
    function userExist(target, value) {
      return  $.ajax({
          type: 'POST',
          url: '/userExists',
          async: false,
          data: {'target': target, 'value': value}
      });
    } 
  }
  
})();