(function() {

  'use strict';
  
  angular
    .module('harbor.project.member')
    .directive('listProjectMember', listProjectMember);
    
  ListProjectMemberController.$inject = ['$scope', 'ListProjectMemberService', '$routeParams'];
    
  function ListProjectMemberController($scope, ListProjectMemberService, $routeParams) {
    var vm = this;
   
    vm.projectId = $routeParams.project_id;
    
    vm.isOpen = false;
    vm.username = "";
            
    vm.search = search; 
    vm.addProjectMember = addProjectMember;
    
    retrieve(vm.projectId , vm.username);
      
    
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
      controller: ListProjectMemberController,
      controllerAs: 'vm',
      bindToController: true
    }
    
    return directive;
  }
  
})();