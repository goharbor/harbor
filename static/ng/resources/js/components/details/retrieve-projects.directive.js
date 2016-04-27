(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .directive('retrieveProjects', retrieveProjects);
  
  RetrieveProjectsController.$inject = ['$scope', 'nameFilter'];
   
  function RetrieveProjectsController($scope, nameFilter) {
    var vm = this;
     
    vm.selectItem = selectItem;
    vm.filterInput = "";
    
    $scope.$watch('vm.selectedProject', function(current, origin) {
      if(current) {        
        var projectId = current.ProjectId;
        vm.selectedId = projectId;     
      }
    });
    
    function selectItem(item) {
      vm.selectedId = item.ProjectId;
      vm.selectedProject = item;
      vm.isOpen = false;
      $scope.$emit('selectedProjectId', vm.selectedId);
    }       
    
  }
  
  function retrieveProjects() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/ng/resources/js/components/details/retrieve-projects.directive.html',
      scope: {
        'projects': '=',
        'isOpen': '=',
        'selectedProject': '='
      },
      replace: true,
      controller: RetrieveProjectsController,
      bindToController: true,
      controllerAs: 'vm'
    }
    
    return directive;
  }
  
})();