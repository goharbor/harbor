(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.account.setting')
    .controller('AccountSettingController', AccountSettingController);
  
  AccountSettingController.$inject = ['CurrentUserService'];
  
  function AccountSettingController(CurrentUserService) {
    var vm = this;
    vm.isOpen = false;
    vm.user = {};
    vm.toggleChangePassword = toggleChangePassword;
    
    CurrentUserService()
      .success(getCurrentUserSuccess)
      .error(getCurrentUserFailed);
  
    vm.update = update;

    function toggleChangePassword() {
      if(vm.isOpen) {
        vm.isOpen = false;
      }else{
        vm.isOpen = true;
      }
      console.log('vm.isOpen:' + vm.isOpen);
    }
    
    function getCurrentUserSuccess(data, status) {
      vm.user = angular.copy(data);
      console.log(data);
    }
    
    function getCurrentUserFailed(data) {
      console.log('Failed get current user:' + data);
    }
    
    function update(user) {
      console.log(user);
    }
    
    
  }
  
})();