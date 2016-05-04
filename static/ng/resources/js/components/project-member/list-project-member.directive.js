(function() {

  'use strict';
  
  angular
    .module('harbor.project.member')
    .directive('listProjectMember', listProjectMember);
    
  ListProjectMemberController.$inject = ['$scope', 'CurrentUserService', 'ListProjectMemberService', '$routeParams'];
    
  function ListProjectMemberController($scope, CurrentUserService, ListProjectMemberService, $routeParams) {
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
      console.log('project_id:' + e.projectId);
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
      $.when(
        CurrentUserService()
          .success(getCurrentUserSuccess)
          .error(getCurrentUserFailed))
      .then(function(){    
        ListProjectMemberService(vm.projectId, {'username': vm.username})
          .then(getProjectMemberComplete)
          .catch(getProjectMemberFailed);      
      });
    }
    
    function getCurrentUserSuccess(data, status) {
      vm.currentUser = data;
    }
    
    function getCurrentUserFailed(e) {
      console.log('Failed in getCurrentUser:' + e);
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