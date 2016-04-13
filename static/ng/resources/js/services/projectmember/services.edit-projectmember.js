(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.projectmember')
    .factory('EditProjectMemberService', EditProjectMemberService);
    
  EditProjectMemberService.$inject = ['$http', '$log'];
  
  function EditProjectMemberService($http, $log) {
    
    return EditProjectMember;
    
    function EditProjectMember(projectMember) {
      
    }
    
  }
  
})();