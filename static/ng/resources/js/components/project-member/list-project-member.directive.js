(function() {

  'use strict';
  
  angular
    .module('harbor.project.member')
    .directive('listProjectMember', listProjectMember);
    
  ListProjectMemberController.$inject = ['$scope', 'ListProjectMemberService', '$routeParams', 'currentUser'];
    
  function ListProjectMemberController($scope, ListProjectMemberService, $routeParams, currentUser) {
    var vm = this;
    
    vm.isOpen = false;      
    vm.search = search; 
    vm.addProjectMember = addProjectMember;
    vm.retrieve = retrieve;
    vm.projectId = $routeParams.project_id;
    vm.username = "";
   
    vm.retrieve();
              
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
      templateUrl: '/static/ng/resources/js/components/project-member/list-project-member.directive.html',
      replace: true,
      scope: true,
      controller: ListProjectMemberController,
      controllerAs: 'vm',
      bindToController: true
    }   
    return directive;
  }
  
})();