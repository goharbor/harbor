(function() {

  'use strict';
  
  angular
    .module('harbor.project.member')
    .controller('AddProjectMemberController', AddProjectMemberController);
    
  AddProjectMemberController.$inject = ['AddProjectMemberService'];
  
  function AddProjectMemberController(AddProjectMemberService) {
    
  }

})();