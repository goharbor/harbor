(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.projectmember')
    .factory('AddProjectMemberService', AddProjectMemberService);
    
  AddProjectMemberService.$inject = ['$http', '$log'];
  
  function AddProjectMemberService($http, $log) {
    
    return AddProjectMember;
    
    function AddProjectMember(projectMember) {
      
    }
    
  }
  
})();