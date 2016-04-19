(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .constant('mockupProjects', mockupProjects)
    .controller('DetailsController', DetailsController);
  
  function mockupProjects() {
    var data = [
      { "id": 1, "name" : "myrepo"},
      { "id": 2, "name" : "myproject"},
      { "id": 3, "name" : "harbor_project"},
      { "id": 4, "name" : "legacy"} 
    ];
    return data;
  }  
  
  DetailsController.$inject = ['mockupProjects', '$scope'];
  
  function DetailsController(mockupProjects, $scope) {
    var vm = this;
    vm.isOpen = false;
    vm.projects = mockupProjects();
    vm.selectedProject = vm.projects[0];
    vm.closeRetrievePane = closeRetrievePane;
    
    function closeRetrievePane() {
      $scope.$broadcast('isOpen', false);
    }
  }
  
})();