(function() {

  'use strict';
  
  angular
    .module('harbor.project.member')
    .directive('listProjectMember', listProjectMember);
    
  ListProjectMemberController.$inject = ['$scope', 'ListProjectMemberService', 'DeleteProjectMemberService', 'getParameterByName', '$location', 'currentUser'];
    
  function ListProjectMemberController($scope, ListProjectMemberService, DeleteProjectMemberService, getParameterByName, $location, currentUser) {
    var vm = this;
    
    vm.isOpen = false;      
    vm.search = search; 
    vm.addProjectMember = addProjectMember;
    vm.deleteProjectMember = deleteProjectMember;
    vm.deleteMember = deleteMember
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
    
    function deleteProjectMember() {
      DeleteProjectMemberService(vm.selectedProjectId, vm.selectedUserId)
        .success(deleteProjectMemberSuccess)
        .error(deleteProjectMemberFailed);
    }
    
    function deleteMember(e) {
      vm.selectedProjectId = e.projectId;
      vm.selectedUserId = e.userId;
      
      vm.modalTitle = 'Delete project member';
      vm.modalMessage = 'Are you sure to delete the current member?';
      
    }
    
    function deleteProjectMemberSuccess(data, status) {
      console.log('Successful delete project member complete.');
      vm.retrieve();      
    }
    
    function deleteProjectMemberFailed(e) {
      console.log('Failed to edit project member:' + e);
    }
  
    function retrieve() {
      ListProjectMemberService(vm.projectId, {'username': vm.username})
        .then(getProjectMemberComplete)
        .catch(getProjectMemberFailed);             
    }
    
    function getProjectMemberComplete(response) {  
      vm.user = currentUser.get();
      vm.projectMembers = response.data;  
    } 
           
    function getProjectMemberFailed(response) {
      console.log('Failed get project members:' + response);
    }
    
  }
  
  function listProjectMember() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/resources/js/components/project-member/list-project-member.directive.html',
      scope: true,
      controller: ListProjectMemberController,
      controllerAs: 'vm',
      bindToController: true
    };
       
    return directive;
  }
  
})();