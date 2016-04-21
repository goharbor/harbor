(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.project.member')
    .factory('EditProjectMemberService', EditProjectMemberService);
    
  EditProjectMemberService.$inject = ['$http', '$log'];
  
  function EditProjectMemberService($http, $log) {
    
    return EditProjectMember;
    
    function EditProjectMember(projectMember) {
      
    }
    
  }
  
})();