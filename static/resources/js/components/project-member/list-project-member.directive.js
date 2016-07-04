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
    
    $scope.$on('$locationChangeSuccess', function() {
      vm.projectId = getParameterByName('project_id', $location.absUrl());
      vm.retrieve();
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
      
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_get_project_member'));
      $scope.$emit('raiseError', true);
            
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
      'controller': ListProjectMemberController,
      'controllerAs': 'vm',
      'bindToController': true
    };
       
    return directive;
  }
  
})();