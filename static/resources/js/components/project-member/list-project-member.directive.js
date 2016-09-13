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
    .module('harbor.project.member')
    .directive('listProjectMember', listProjectMember);
    
  ListProjectMemberController.$inject = ['$scope', 'ListProjectMemberService', 'DeleteProjectMemberService', 'getParameterByName', '$location', 'currentUser', '$filter', 'trFilter', '$window'];
    
  function ListProjectMemberController($scope, ListProjectMemberService, DeleteProjectMemberService, getParameterByName, $location, currentUser, $filter, trFilter, $window) {
    
    $scope.subsTabPane = 30;
    
    var vm = this;

    vm.sectionHeight = {'min-height': '579px'};
    
    vm.isOpen = false;      
    vm.search = search; 
    vm.addProjectMember = addProjectMember;
    vm.deleteProjectMember = deleteProjectMember;
    vm.retrieve = retrieve;
    vm.username = '';
    
    vm.projectId = getParameterByName('project_id', $location.absUrl());
    vm.retrieve();
    
    $scope.$on('retrieveData', function(e, val) {
      if(val) {
        console.log('received retrieve data:' + val);
        vm.projectId = getParameterByName('project_id', $location.absUrl());
        vm.username = '';
        vm.retrieve();
      }
    });
              
    function search(e) {
      vm.projectId = e.projectId;
      vm.username = e.username;
      retrieve();
    }
    
    function addProjectMember() {
      if(vm.isOpen) {
        vm.isOpen = false;
      }else{
        vm.isOpen = true;
      }
    }
    
    function deleteProjectMember(e) {
      DeleteProjectMemberService(e.projectId, e.userId)
        .success(deleteProjectMemberSuccess)
        .error(deleteProjectMemberFailed);
    }
        
    function deleteProjectMemberSuccess(data, status) {
      console.log('Successful delete project member.');
      vm.retrieve();      
    }
    
    function deleteProjectMemberFailed(e) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_delete_member'));
      $scope.$emit('raiseError', true);
      console.log('Failed to edit project member:' + e);
    }
  
    function retrieve() {
      ListProjectMemberService(vm.projectId, {'username': vm.username})
        .then(getProjectMemberComplete)
        .catch(getProjectMemberFailed);             
    }
    
    function getProjectMemberComplete(response) {  
      vm.user = currentUser.get();
      vm.projectMembers = response.data || [];  
    } 
           
    function getProjectMemberFailed(response) {
      console.log('Failed to get project members:' + response);
      vm.projectMembers = [];    
      $location.url('repositories').search('project_id', vm.projectId);
    }
    
  }
  
  function listProjectMember() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/project-member/list-project-member.directive.html',
      'scope': {
        'sectionHeight': '='
      },
      'link': link,
      'controller': ListProjectMemberController,
      'controllerAs': 'vm',
      'bindToController': true
    };
       
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      element.find('#txtSearchInput').on('keydown', function(e) {
        if($(this).is(':focus') && e.keyCode === 13) {
          ctrl.retrieve();
        }
      });
    }
  }
  
})();