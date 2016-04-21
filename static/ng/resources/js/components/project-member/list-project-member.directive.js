(function() {

  'use strict';
  
  angular
    .module('harbor.project.member')
    .constant('mockupProjectMembers', mockupProjectMembers)
    .directive('listProjectMember', listProjectMember);
    
  function mockupProjectMembers() {
    var projectMembers = [
      {'id': '1', 'username': 'user1', 'roleId': '1'},
      {'id': '2', 'username': 'user2', 'roleId': '3'},
      {'id': '3', 'username': 'user3', 'roleId': '2'}
    ];
    return projectMembers;
  }
  
  ListProjectMemberController.$inject = ['mockupProjectMembers', 'ListProjectMemberService'];
    
  function ListProjectMemberController(mockupProjectMembers, ListProjectMemberService) {
    var vm = this;
    
    vm.isOpen = false;
    vm.addMember = addMember;
    
    function addMember() {
      if(vm.isOpen) {
        vm.isOpen = false;
      }else{
        vm.isOpen = true;
      }
    }
    
    vm.projectMembers = mockupProjectMembers();
    
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