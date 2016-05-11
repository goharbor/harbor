(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .directive('retrieveProjects', retrieveProjects);
  
  RetrieveProjectsController.$inject = ['$scope', 'nameFilter', '$filter', 'CurrentProjectMemberService', 'ListProjectService', '$routeParams', '$location'];
   
  function RetrieveProjectsController($scope, nameFilter, $filter, CurrentProjectMemberService, ListProjectService, $routeParams, $location) {
    var vm = this;
   
    vm.projectName = '';
    vm.isPublic = 0;        
    vm.publicity = false;
    
    vm.retrieve = retrieve;
    
    vm.retrieve();
    
    vm.checkProjectMember = checkProjectMember;
    
    $scope.$watch('vm.selectedProject', function(current, origin) {
      if(current) {        
        vm.selectedId = current.ProjectId;
      }
    });
    
    vm.filterInput = "";
    vm.selectItem = selectItem;  
    
    $scope.$watch('vm.publicity', function(current, origin) { 
      vm.isPublic = current ? 1 : 0;        
      vm.retrieve();      
    });
       
    function retrieve() {
      ListProjectService(vm.projectName, vm.isPublic)
        .success(getProjectSuccess)
        .error(getProjectFailed);
    }
    
    function getProjectSuccess(data, status) {
      vm.projects = data;
      if(vm.projects == null) {
        vm.publicity = 1;
        console.log('vm.projects is null, load public projects.');
      }
      
      vm.selectedProject = vm.projects[0];  
      
      if($routeParams.project_id){
        angular.forEach(vm.projects, function(value, index) {
          if(value['ProjectId'] == $routeParams.project_id) {
            vm.selectedProject = value;
          }
        }); 
      }
      
      $location.search('project_id', vm.selectedProject.ProjectId);
      vm.checkProjectMember(vm.selectedProject.ProjectId);
      vm.resultCount = vm.projects.length;
    
      $scope.$watch('vm.filterInput', function(current, origin) {  
        vm.resultCount = $filter('name')(vm.projects, vm.filterInput, 'Name').length;
      });
    }
    
    function getProjectFailed(response) {
      console.log('Failed to list projects:' + response);
    }
    
    function checkProjectMember(projectId) {
      CurrentProjectMemberService(projectId)
        .success(getCurrentProjectMemberSuccess)
        .error(getCurrentProjectMemberFailed);
    }
    
    function getCurrentProjectMemberSuccess(data, status) {
      console.log('Successful get current project member:' + status);
      vm.isProjectMember = true;
    }
    
    function getCurrentProjectMemberFailed(data, status) {
      console.log('Failed get current project member:' + status);
      vm.isProjectMember = false;
    }
  
    function selectItem(item) {
      vm.selectedId = item.ProjectId;
      vm.selectedProject = item;
      vm.checkProjectMember(vm.selectedProject.ProjectId);
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
        'selectedProject': '=',
        'publicity': '=',
        'isProjectMember': '='
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