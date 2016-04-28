(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .controller('DetailsController', DetailsController);
    
  DetailsController.$inject = ['ListProjectService', '$scope'];
  
  function DetailsController(ListProjectService, $scope) {
    var vm = this;
    vm.isOpen = false;
    vm.closeRetrievePane = closeRetrievePane;
    
    $scope.$on('selectedProjectId', function(e, val) {
      $scope.$broadcast('currentProjectId', val);
    });    
    
    ListProjectService({'isPublic' : 0, 'projectName' : ''})
      .then(getProjectComplete)
      .catch(getProjectFailed);
      
    function getProjectComplete(response) {
      vm.projects = response.data;
      vm.selectedProject = vm.projects[0];
    }
    
    function getProjectFailed(response) {
      
    }
    
    function closeRetrievePane() {
      $scope.$broadcast('isOpen', false);
    }
  }
  
})();