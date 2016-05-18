(function() {
  
  angular
    .module('harbor.app')
    .factory('currentUser', currentUser)
    .factory('projectMember', projectMember);
    
  function currentUser() {
    var currentUser;
    return {
      set: function(user) {
        currentUser = user;
        console.log('set currentUser:' + currentUser);
      },
      get: function() {
        console.log('get currentUser:' + currentUser);
        return currentUser;
      }
    }
  }  
  
  function projectMember() {
    var projectMember;
    return {
      set: function(member) {
        projectMember = member;
        console.log('set projectMember:');
        console.log(projectMember);
      },
      get: function() {
        console.log('get projectMember:');
        console.log(projectMember);
        return projectMember;
      }
    }
  }
      
})();