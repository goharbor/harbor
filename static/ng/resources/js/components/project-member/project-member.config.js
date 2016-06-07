(function() {
  
  'use strict';
  
  angular
    .module('harbor.project.member')
    .constant('roles', roles)
    .factory('getRole', getRole);
    
  function roles() {
    return [
      {'id': '0', 'name': 'NA', 'roleName': 'NA'},
      {'id': '1', 'name': 'Project Admin', 'roleName': 'projectAdmin'},
      {'id': '2', 'name': 'Developer', 'roleName': 'developer'},
      {'id': '3', 'name': 'Guest', 'roleName': 'guest'}
    ];
  }
  
  getRole.$inject = ['roles'];
  
  function getRole(roles) {
    var r = roles();
    return get;     
    function get(query) {
     
      for(var i = 0; i < r.length; i++) {
        var role = r[i];
        if(query.key === 'roleName' && role.roleName === query.value
          || query.key === 'roleId' && role.id === String(query.value)) {
           return role;
        }
      }
    }
  }
})();