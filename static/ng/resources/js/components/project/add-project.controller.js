(function() {
  
  'use strict';
  
  angular
    .module('harbor.project')
    .controller('AddProjectController', AddProjectController);
    
  AddProjectController.$inject = ['AddProjectService'];
  
  function AddProjectController(AddProjectService) {
    
  }
   
})();