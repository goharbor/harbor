(function() {
  'use strict';
  
  angular
    .module('harbor.services.project')
    .factory('AddProjectService', AddProjectService);
    
  AddProjectService.$inject = ['$http', '$log'];
    
  function AddProjectService($http, $log) {
    
    return AddProject;
    
    function AddProject(project) {
       
    }
  }
  
})();