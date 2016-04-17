(function() {

  'use strict';
  
  angular
    .module('harbor.user')
    .directive('listUser', listUser);
    
  ListUserController.$inject = ['ListUserService'];
    
  function ListUserController(ListUserService) {
    
  }
  
  function listUser() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/ng/resources/js/components/user/list-user.directive.html',
      replace: true,
      controller: ListUserController,
      controllerAs: 'vm',
      bindToController: true
    }
    
    return directive;
  }
  
})();