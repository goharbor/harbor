(function() {
  
  'use strict';
  
   angular
    .module('harbor.services.project.member')
    .service('ListProjectMemberService', ListProjectMemberService);
   
  ListProjectMemberService.$inject = ['$http', '$log'];
 
  function ListProjectMemberService() {
    
    return ListProjectMember;
    
    function ListProjectMember () {
      
    }
  }
  
})();