(function() {

  'use strict';
  
  angular
    .module('harbor.projectmember')
    .controller('AddProjectMemberController', AddProjectMemberController);
    
  AddProjectMemberController.$inject = ['AddProjectMemberService'];
  
  function AddProjectMemberController(AddProjectMemberService) {
    
  }

})();