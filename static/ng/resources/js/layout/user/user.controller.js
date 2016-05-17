(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.user')
    .controller('UserController', UserController);
    
  UserController.$inject = ['ListUserService', 'DeleteUserService'];
  
  function UserController(ListUserService, DeleteUserService) {
    var vm = this;
    
    vm.username = '';
    vm.searchUser = searchUser;
    vm.deleteUser = deleteUser;
    vm.retrieve = retrieve;
    
    vm.retrieve();
    
    function searchUser() {
      vm.retrieve();
    }
    
    function deleteUser(userId) {
      DeleteUserService(userId)
        .success(deleteUserSuccess)
        .error(deleteUserFailed);
    }
    
    function retrieve() {
      ListUserService(vm.username)
        .success(listUserSuccess)
        .error(listUserFailed);
    }
    
    function deleteUserSuccess(data, status) {
      console.log('Successful delete user.');
      vm.retrieve();
    }
    
    function deleteUserFailed(data, status) {
      console.log('Failed delete user.');
    }
    
    function listUserSuccess(data, status) {
      vm.users = data;
    }
    
    function listUserFailed(data, status) {
      console.log('Failed list user:' + data);
    }
    
    
      
  }
  
})();