(function() {

  'use strict';
  
  angular
    .module('harbor.project.member')
    .directive('listProjectMember', listProjectMember);
    
  ListProjectMemberController.$inject = ['ListProjectMemberService', '$routeParams'];
    
  function ListProjectMemberController(ListProjectMemberService, $routeParams) {
    var vm = this;
    
    vm.isOpen = false;
    vm.username = "";
    
    vm.search = search; 
    vm.addProjectMember = addProjectMember;
    
    vm.projectId = $routeParams.project_id || 2;
    
    retrieve(vm.username);
    
    function search(e) {
      retrieve(e.username);
    }
    
    function addProjectMember() {
      if(vm.isOpen) {
        vm.isOpen = false;
      }else{
        vm.isOpen = true;
      }
    }
    
    function retrieve(username) {    
      ListProjectMemberService(vm.projectId, {'username': username})
        .then(getProjectMemberComplete)
        .catch(getProjectMemberFailed);        
    }
    
    function getProjectMemberComplete(response) {
      vm.projectMembers = response.data;  
    } 
           
    function getProjectMemberFailed(response) {
      
    }
    
  }
  
  function listProjectMember() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/ng/resources/js/components/project-member/list-project-member.directive.html',
      replace: true,
      controller: ListProjectMemberController,
      controllerAs: 'vm',
      bindToController: true
    }
    
    return directive;
  }
  
})();