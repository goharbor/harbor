/*
    Copyright (c) 2016 VMware, Inc. All Rights Reserved.
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
        
        http://www.apache.org/licenses/LICENSE-2.0
        
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/
(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .directive('retrieveProjects', retrieveProjects);
  
  RetrieveProjectsController.$inject = ['$scope', 'nameFilter', '$filter', 'trFilter', 'ListProjectService', '$location', 'getParameterByName', 'CurrentProjectMemberService', '$window'];
   
  function RetrieveProjectsController($scope, nameFilter, $filter, trFilter, ListProjectService, $location, getParameterByName, CurrentProjectMemberService, $window) {
    var vm = this;
    
    vm.projectName = '';
    vm.isOpen = false;
    vm.isProjectMember = false;
    vm.target = $location.path().substr(1) || 'repositories';
    
    vm.isPublic = Number(getParameterByName('is_public', $location.absUrl()));
   
    var DEFAULT_PAGE = 1;
        
    vm.page = DEFAULT_PAGE;
       
    vm.projects = [];
    vm.retrieve = retrieve;
    vm.filterInput = '';
    vm.selectItem = selectItem;  
    vm.checkProjectMember = checkProjectMember;  
                     
    function retrieve() {
      ListProjectService(vm.projectName, vm.isPublic, vm.page)
        .then(getProjectSuccess, getProjectFailed);
    }
    
    $scope.$watch('vm.page', function(current) {
      if(current) {
        vm.retrieve();
      }
    });
    
    $scope.$watch('vm.isPublic', function(current) {
      vm.projectType = vm.isPublic === 0 ? 'my_project_count' : 'public_project_count';
    });
    
    $scope.$watch('vm.selectedProject', function(current) {
      if(current) {
        vm.selectedId = current.project_id;
      }
    });
    
    function parseNextLink(link) {
      if(link === '') {
        return false;
      }
      
      var parts = link.split(",");
      for(var i in parts) {
        var groups = /^\<(.*)\>;\srel=\"(\w+)\"$/.exec($.trim(parts[i]));
        if(groups && groups.length > 2){
          var url = groups[1];
          var rel = groups[2];
          if(rel === 'next') {
            return {
              'page': getParameterByName('page', url),
              'rel' : rel
            };
          }
        }
      }
      return false;
    }
    
    
    function getProjectSuccess(response) {
      
      var partialProjects = response.data || []; 
      for(var i in partialProjects) {
        vm.projects.push(partialProjects[i]);
      }
      
      var nextLink = parseNextLink(response.headers("Link") || '');
      
      if(nextLink) {
        vm.page = parseInt(nextLink.page);
      } else {
        
        if(vm.projects.length == 0 && vm.isPublic === 0){
          $window.location.href = '/project';  
        }
                  
        if(getParameterByName('project_id', $location.absUrl())){
          for(var i in vm.projects) {
            var project = vm.projects[i];
            if(project['project_id'] == getParameterByName('project_id', $location.absUrl())) {
              vm.selectedProject = project;
              break;
            }
          } 
        }
  
        $location.search('project_id', vm.selectedProject.project_id);
        vm.checkProjectMember(vm.selectedProject.project_id);         
           
        vm.resultCount = vm.projects.length;
      
        $scope.$watch('vm.filterInput', function(current, origin) {  
          vm.resultCount = $filter('name')(vm.projects, vm.filterInput, 'name').length;
        });
      }
    }
    
    function getProjectFailed(response) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_get_project'));
      $scope.$emit('raiseError', true);
      console.log('Failed to list projects.');
    }
      
    function selectItem(item) {
      vm.selectedProject = item;
      $location.search('project_id', vm.selectedProject.project_id);
      $scope.$emit('projectChanged', true);
    }       
  
    $scope.$on('$locationChangeSuccess', function(e) {
      vm.projectId = getParameterByName('project_id', $location.absUrl());
      vm.isOpen = false;
      vm.checkProjectMember(vm.selectedProject.project_id);
    });
    
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
      vm.isProjectMember = false;  
      console.log('Current user has no member for the project:' + status +  ', location.url:' + $location.url());
      vm.target = 'repositories';
    }
    
  }
  
  function retrieveProjects() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/resources/js/components/details/retrieve-projects.directive.html',
      scope: {
        'target': '=',
        'isOpen': '=',
        'selectedProject': '=',
        'isPublic': '=',
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
          if (!$(this).is(e.target) && 
               $(this).has(e.target).length === 0 &&
               $('.popover').has(e.target).length === 0) {
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