(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.user')
    .factory('UpdateUserService', UpdateUserService);
    
  UpdateUserService.$inject = ['$http'];
  
  function UpdateUserService($http) {
    return updateUser;
    function updateUser(userId, user) {
      return $http
        .put('/api/users/' + userId, {
          'username': user.username,
          'email': user.email,
          'realname': user.realname,
          'comment': user.comment
        });
    }
  }
  
})();