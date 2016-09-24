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
    .module('harbor.layout.project')
    .controller('ProjectController', ProjectController);

  ProjectController.$inject = ['$scope', 'ListProjectService', 'DeleteProjectService', '$timeout', 'currentUser', 'getRole', '$filter', 'trFilter', 'getParameterByName', '$location']; 

  function ProjectController($scope, ListProjectService, DeleteProjectService, $timeout, currentUser, getRole, $filter, trFilter, getParameterByName, $location) {
    var vm = this;
 
    vm.isOpen = false;
    vm.projectName = '';
    vm.isPublic = Number(getParameterByName('is_public', $location.absUrl())) || 0;
    
    vm.page = 1;
    vm.pageSize = 15;  
     
    vm.sectionHeight = {'min-height': '579px'};    
    
    vm.retrieve = retrieve;
    vm.showAddProject = showAddProject;
    vm.searchProject = searchProject;    
    vm.showAddButton = showAddButton;
    vm.togglePublicity = togglePublicity;    
    vm.user = currentUser.get();      
    vm.getProjectRole = getProjectRole;
    
    vm.searchProjectByKeyPress = searchProjectByKeyPress;
    vm.confirmToDelete = confirmToDelete;
    vm.deleteProject = deleteProject;
    
   
    
    //Error message dialog handler for project.
    $scope.$on('modalTitle', function(e, val) {
      vm.modalTitle = val;
    });
    
    $scope.$on('modalMessage', function(e, val) {
      vm.modalMessage = val;
    });
       
    $scope.$on('raiseError', function(e, val) {
      if(val) {   
        vm.action = function() {
          $scope.$broadcast('showDialog', false);
        };
        vm.contentType = 'text/plain';
        vm.confirmOnly = true;      
        $timeout(function() {
          $scope.$broadcast('showDialog', true);
        }, 350);
      }
    });
    
    $scope.$on('raiseInfo', function(e, val) {
      if(val) {
        vm.action = function() {
          val.action();
          $scope.$broadcast('showDialog', false);
        };
        vm.contentType = val.contentType;
        vm.confirmOnly = val.confirmOnly;
       
        $scope.$broadcast('showDialog', true);
      }
    });
    
    
    $scope.$watch('vm.page', function(current) {
      if(current) {
        vm.page = current;
        vm.retrieve();
      }
    });
    
    function retrieve() {       
      ListProjectService(vm.projectName, vm.isPublic, vm.page, vm.pageSize)
        .then(listProjectSuccess)
        .catch(listProjectFailed);
    }
    
    function listProjectSuccess(response) {
      vm.totalCount = response.headers('X-Total-Count');
      vm.projects = response.data || [];
    }
    
    function getProjectRole(roleId) {
      if(roleId !== 0) {
        var role = getRole({'key': 'roleId', 'value': roleId});
        return role.name;
      }
      return '';
    }
    
    function listProjectFailed(response) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_get_project'));
      $scope.$emit('raiseError', true);
      console.log('Failed to get Project.');
    }
          
    $scope.$on('addedSuccess', function(e, val) {
      vm.retrieve();
    });
   
    function showAddProject() {
      vm.isOpen = vm.isOpen ? false : true;
    }
    
    function searchProject() {
      vm.retrieve();
    }
    
    function showAddButton() {
      return (vm.isPublic === 0);
    }
    
    function togglePublicity(e) {
      vm.isPublic = e.isPublic;
      vm.isOpen = false;
      vm.page = 1;
      vm.retrieve();
    }
    
    function searchProjectByKeyPress($event) {
      var keyCode = $event.which || $event.keyCode;
      if(keyCode === 13) {
        vm.retrieve();
      }
    }
    
    function confirmToDelete(projectId, projectName) {
      vm.selectedProjectId = projectId;
     
      $scope.$emit('modalTitle', $filter('tr')('confirm_delete_project_title'));
      $scope.$emit('modalMessage', $filter('tr')('confirm_delete_project', [projectName]));
      
      var emitInfo = {
        'confirmOnly': false,
        'contentType': 'text/plain',
        'action': vm.deleteProject
      };
      
      $scope.$emit('raiseInfo', emitInfo);
    }
    
    function deleteProject() {
      DeleteProjectService(vm.selectedProjectId)
        .success(deleteProjectSuccess)
        .error(deleteProjectFailed);
    }
    
    function deleteProjectSuccess(data, status) {
      console.log('Successful delete project.');
      vm.retrieve();
    }
    
    function deleteProjectFailed(data, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      if(status === 412) {
        $scope.$emit('modalMessage', $filter('tr')('failed_to_delete_project_contains_repo'));
      }
      if(status === 403) {
        $scope.$emit('modalMessage', $filter('tr')('failed_to_delete_project_insuffient_permissions'));  
      }
      $scope.$emit('raiseError', true);
      console.log('Failed to delete project.');
    }
    
  }
  
})();