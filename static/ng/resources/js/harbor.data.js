(function() {
  
  'use strict';
  
  angular
    .module('harbor.app') 
    .factory('currentUser', currentUser)
    .factory('currentProjectMember', currentProjectMember);
  
  currentUser.$inject = ['$cookies', '$timeout'];
  
  function currentUser($cookies, $timeout) {
    return {
      set: function(user) {
        $cookies.putObject('user', user, {'path': '/'});
      },
      get: function() {
        return $cookies.getObject('user');
      },
      unset: function() {
        $cookies.remove('user', {'path': '/'});
      }
    }
  }  
  
  currentProjectMember.$inject = ['$cookies'];
  
  function currentProjectMember($cookies) {
    return {
      set: function(member) {
        $cookies.putObject('member', member, {'path': '/'});
      },
      get: function() {
        return $cookies.getObject('member');
      },
      unset: function() {
        $cookies.remove('member', {'path': '/'});
      }
    }
  }
      
})();