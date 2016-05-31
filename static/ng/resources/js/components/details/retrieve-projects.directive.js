(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .directive('retrieveProjects', retrieveProjects);
  
  RetrieveProjectsController.$inject = ['$scope', 'nameFilter', '$filter', 'CurrentProjectMemberService', 'ListProjectService', '$location', 'getParameterByName', '$window'];
   
  function RetrieveProjectsController($scope, nameFilter, $filter, CurrentProjectMemberService, ListProjectService, $location, getParameterByName, $window) {
    var vm = this;
    
    vm.projectName = '';
    vm.isOpen = false;
    
    if(getParameterByName('is_public', $location.absUrl())) {
      vm.isPublic = getParameterByName('is_public', $location.absUrl()) === 'true' ? 1 : 0;
      vm.publicity = (vm.isPublic === 1) ? true : false;
    }

    vm.retrieve = retrieve;
    vm.checkProjectMember = checkProjectMember;
    
    $scope.$watch('vm.selectedProject', function(current, origin) {
      if(current) {        
        vm.selectedId = current.ProjectId;
      }
    });
    
    vm.filterInput = "";
    vm.selectItem = selectItem;  
    
    $scope.$watch('vm.publicity', function(current, origin) { 
      vm.publicity = current ? true : false;
      vm.isPublic =  vm.publicity ? 1 : 0;
      vm.projectType = (vm.isPublic === 1) ? 'public_projects' : 'my_projects';
      vm.retrieve();      
    });
       
    function retrieve() {
      ListProjectService(vm.projectName, vm.isPublic)
        .success(getProjectSuccess)
        .error(getProjectFailed);
    }
    
    function getProjectSuccess(data, status) {
      vm.projects = data;

      if(!angular.isDefined(vm.projects)) {
        vm.isPublic = 1;
        vm.publicity = 1;
        vm.projectType = 'public_projects';
        console.log('vm.projects is undefined, load public projects.');
      }
      
      vm.selectedProject = vm.projects[0];
      
      if(getParameterByName('project_id', $location.absUrl())){
        angular.forEach(vm.projects, function(value, index) {
          if(value['ProjectId'] === Number(getParameterByName('project_id', $location.absUrl()))) {
            vm.selectedProject = value;
          }
        }); 
      }
      vm.checkProjectMember(vm.selectedProject.ProjectId);
      $location.search('project_id', vm.selectedProject.ProjectId);
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
      $window.location.reload();
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
      controller: RetrieveProjectsController,
      bindToController: true,
      controllerAs: 'vm'
    };
    
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      $(document).on('click', clickHandler);
    
      function clickHandler(e) {
        $('[data-toggle="popover"]').each(function () {          
          //the 'is' for buttons that trigger popups
          //the 'has' for icons within a button that triggers a popup
          if (!$(this).is(e.target) && $(this).has(e.target).length === 0 && $('.popover').has(e.target).length === 0) {
            $(this).parent().popover('hide');
          }
        });
        var targetId = $(e.target).attr('id');
        if(targetId === 'switchPane' || 
           targetId === 'retrievePane' ||
           targetId === 'retrieveFilter') {
          return;            
        }else{
          ctrl.isOpen = false;
          scope.$apply();
        }
      }
    }

  }
  
})();