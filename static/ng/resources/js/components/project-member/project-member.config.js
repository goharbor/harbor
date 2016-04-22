(function() {
  
  'use strict';
  
  angular
    .module('harbor.project.member')
    .constant('roles', roles)
    .factory('getRoleById', getRoleById);
    
  function roles() {
    return [
      {'id': '1', 'name': 'Project Admin'},
      {'id': '2', 'name': 'Developer'},
      {'id': '3', 'name': 'Guest'}
    ];
  }
  
  getRoleById.$inject = ['roles'];
  
  function getRoleById(roles) {
    var r = roles();
    return getRole;     
    function getRole(roleId) {
     
      for(var i = 0; i < r.length; i++) {
        var role = r[i];
        if(role.id == roleId) {
          return role;
        }
      }
    }
  }
})();