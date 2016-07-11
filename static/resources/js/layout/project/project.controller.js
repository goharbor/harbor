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

  ProjectController.$inject = ['$scope', 'ListProjectService', '$timeout', 'currentUser', 'getRole', '$filter', 'trFilter']; 

  function ProjectController($scope, ListProjectService, $timeout, currentUser, getRole, $filter, trFilter) {
    var vm = this;
 
    vm.isOpen = false;
    vm.projectName = '';
    vm.publicity = 0;
     
    vm.retrieve = retrieve;
    vm.showAddProject = showAddProject;
    vm.searchProject = searchProject;    
    vm.showAddButton = showAddButton;
    vm.togglePublicity = togglePublicity;    
    vm.user = currentUser.get();      
    vm.retrieve();
    vm.getProjectRole = getProjectRole;
    
    vm.searchProjectByKeyPress = searchProjectByKeyPress;
    
    
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
        $scope.$broadcast('showDialog', true);
      }
    });
    
    
    function retrieve() {       
      ListProjectService(vm.projectName, vm.publicity)
        .success(listProjectSuccess)
        .error(listProjectFailed);
    }
    
    function listProjectSuccess(data, status) {
      vm.projects = data || [];
    }
    
    function getProjectRole(roleId) {
      if(roleId !== 0) {
        var role = getRole({'key': 'roleId', 'value': roleId});
        return role.name;
      }
      return '';
    }
    
    function listProjectFailed(data, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_get_project'));
      $scope.$emit('raiseError', true);
      console.log('Failed to get Project.');
    }
          
    $scope.$on('addedSuccess', function(e, val) {
      vm.retrieve();
    });
   
    function showAddProject() {
      if(vm.isOpen){
        vm.isOpen = false;        
      }else{
        vm.isOpen = true;        
      }
    }
    
    function searchProject() {
      vm.retrieve();
    }
    
    function showAddButton() {
      if(vm.publicity === 0) {
        return true;
      }else{
        return false;
      }
    }
    
    function togglePublicity(e) {
      vm.publicity = e.publicity;
      vm.isOpen = false;
      vm.retrieve();
      console.log('vm.publicity:' + vm.publicity);
    }
    
    function searchProjectByKeyPress($event) {
      var keyCode = $event.which || $event.keyCode;
      if(keyCode === 13) {
        vm.retrieve();
      }
    }
    
  }
  
})();