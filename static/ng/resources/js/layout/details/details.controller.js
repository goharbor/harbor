(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .controller('DetailsController', DetailsController);
    
  DetailsController.$inject = ['ListProjectService', '$scope', '$location'];
  
  function DetailsController(ListProjectService, $scope, $location) {
    var vm = this;
    vm.isOpen = false;
    vm.closeRetrievePane = closeRetrievePane;   
    
    ListProjectService({'isPublic' : 0, 'projectName' : ''})
      .then(getProjectComplete)
      .catch(getProjectFailed);
      
    function getProjectComplete(response) {
      vm.projects = response.data;
      vm.selectedProject = vm.projects[0];
      $location.url('repositories').search('project_id', vm.selectedProject.ProjectId);
    }
    
    function getProjectFailed(response) {
      
    }
    
    function closeRetrievePane() {
      $scope.$broadcast('isOpen', false);
    }
  }
  
})();