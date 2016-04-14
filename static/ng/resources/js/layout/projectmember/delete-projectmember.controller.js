(function() {

  'use strict';
  
  angular
    .module('harbor.projectmember')
    .controller('DeleteProjectMemberController', DeleteProjectMemberController);
    
  DeleteProjectMemberController.$inject = ['DeleteProjectMemberService'];
  
  function DeleteProjectMemberController(DeleteProjectMemberService) {
    
  }

})();