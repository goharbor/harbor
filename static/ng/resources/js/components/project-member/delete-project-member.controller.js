(function() {

  'use strict';
  
  angular
    .module('harbor.project.member')
    .controller('DeleteProjectMemberController', DeleteProjectMemberController);
    
  DeleteProjectMemberController.$inject = ['DeleteProjectMemberService'];
  
  function DeleteProjectMemberController(DeleteProjectMemberService) {
    
  }

})();