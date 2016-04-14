(function() {

  'use strict';
  
  angular
    .module('harbor.project')
    .controller('ListProjectController', ListProjectController);
  
  ListProjectController.$inject = ['ListProjectService'];
  
  function ListProjectService(ListProjectService) {
    
  }
  
})();