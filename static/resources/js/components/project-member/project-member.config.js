(function() {
  
  'use strict';
  
  angular
    .module('harbor.project.member')
    .constant('roles', roles)
    .factory('getRole', getRole);
    
  function roles() {
    return [
      {'id': '1', 'name': 'project_admin', 'roleName': 'projectAdmin'},
      {'id': '2', 'name': 'developer', 'roleName': 'developer'},
      {'id': '3', 'name': 'guest', 'roleName': 'guest'}
    ];
  }
  
  getRole.$inject = ['roles', '$filter', 'trFilter'];
  
  function getRole(roles, $filter, trFilter) {
    var r = roles();
    return get;     
    function get(query) {
     
      for(var i = 0; i < r.length; i++) {
        var role = r[i];
        if(query.key === 'roleName' && role.roleName === query.value
          || query.key === 'roleId' && role.id === String(query.value)) {
            console.log('role.name: ' + role.name);
           return role;
        }
      }
    }
  }
})();