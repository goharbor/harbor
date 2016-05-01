(function() {

  'use strict';
  
  angular
    .module('harbor.project.member')
    .directive('listProjectMember', listProjectMember);
    
  ListProjectMemberController.$inject = ['$scope', '$q', 'ListProjectMemberService', '$routeParams'];
    
  function ListProjectMemberController($scope, $q, ListProjectMemberService, $routeParams) {
    var vm = this;
    
    vm.isOpen = false;
    vm.username = "";
            
    vm.search = search; 
    vm.addProjectMember = addProjectMember;
    vm.retrieve = retrieve;
    
    vm.projectId = $routeParams.project_id;
    vm.retrieve(vm.projectId, vm.username);
              
    function search(e) {
      console.log('project_id:' + e.projectId);
      retrieve(e.projectId, e.username);
    }
    
    function addProjectMember() {
      if(vm.isOpen) {
        vm.isOpen = false;
      }else{
        vm.isOpen = true;
      }
    }
    
    function retrieve(projectId, username) {    
      ListProjectMemberService(projectId, {'username': username})
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
      link: link,
      controller: ListProjectMemberController,
      controllerAs: 'vm',
      bindToController: true
    }
   
    return directive;
    
    function link(scope, element, attrs, ctrl) {
     
    }
  }
  
})();