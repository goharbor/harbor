(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.project.member')
    .factory('DeleteProjectMemberService', DeleteProjectMemberService);
    
  DeleteProjectMemberService.$inject = ['$http', '$log'];
  
  function DeleteProjectMemberService($http, $log) {
    
    return DeleteProjectMember;
    
    function DeleteProjectMember(projectMember) {
      
    }
    
  }
  
})();