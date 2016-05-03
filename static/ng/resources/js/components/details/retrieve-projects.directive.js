(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .directive('retrieveProjects', retrieveProjects);
  
  RetrieveProjectsController.$inject = ['$scope', 'nameFilter', '$filter', 'ListProjectService', '$routeParams', '$location'];
   
  function RetrieveProjectsController($scope, nameFilter, $filter, ListProjectService, $routeParams, $location) {
    var vm = this;
   
    vm.projectName = '';
    vm.isPublic = 0;        
    
    ListProjectService(vm.projectName, vm.isPublic)
      .success(getProjectSuccess)
      .error(getProjectFailed);
          
    function getProjectSuccess(data, status) {
      vm.projects = data;
      
      if($routeParams.project_id){
        angular.forEach(vm.projects, function(value, index) {
          if(value['ProjectId'] == $routeParams.project_id) {
            vm.selectedProject = value;
          }
        });
      }else{
        vm.selectedProject = vm.projects[0];  
      }
      vm.resultCount = vm.projects.length;
    
      $scope.$watch('vm.filterInput', function(current, origin) {  
        vm.resultCount = $filter('name')(vm.projects, vm.filterInput, 'Name').length;
      });
      
    }
    
    function getProjectFailed(response) {
      console.log('Failed to list projects:' + response);
    }
  
    $scope.$watch('vm.selectedProject', function(current, origin) {
      if(current) {        
        vm.selectedId = current.ProjectId;
      }
    });
  
    vm.filterInput = "";
    vm.selectItem = selectItem;  
    
    
    function selectItem(item) {
      vm.selectedId = item.ProjectId;
      vm.selectedProject = item;
      vm.isOpen = false;
      $location.search('project_id', vm.selectedProject.ProjectId);
    }       
    
  }
  
  function retrieveProjects() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/ng/resources/js/components/details/retrieve-projects.directive.html',
      scope: {
        'isOpen': '=',
        'selectedProject': '='
      },
      link: link,
      replace: true,
      controller: RetrieveProjectsController,
      bindToController: true,
      controllerAs: 'vm'
    }
    
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      
    }
  }
  
})();