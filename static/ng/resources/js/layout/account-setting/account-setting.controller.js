(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.account.setting')
    .controller('AccountSettingController', AccountSettingController);
  
  AccountSettingController.$inject = ['CurrentUserService', 'ChangePasswordService', '$window'];
  
  function AccountSettingController(CurrentUserService, ChangePasswordService, $window) {
    var vm = this;
    vm.isOpen = false;
    vm.user = {};
    vm.toggleChangePassword = toggleChangePassword;
    
    vm.changeProfile = changeProfile;
    vm.changePassword= changePassword;
    vm.cancel = cancel;
    
    CurrentUserService()
      .success(getCurrentUserSuccess)
      .error(getCurrentUserFailed);

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
    
    function changeProfile(user) {
      console.log(user);
    }
    
    function changePassword(user) {
      console.log(user);
      ChangePasswordService(vm.user.UserId, user.oldPassword, user.password)
        .success(changePasswordSuccess)
        .error(changePasswordFailed);
    }
    
    function changePasswordSuccess(data, status) {
      console.log('Successful changed password:' + data);
      $window.location.href = '/ng/project';
    }
    
    function changePasswordFailed(data, status) {
      console.log('Failed changed password:' + data);
      if(data === 'old_password_is_not_correct') {
        vm.oldPasswordIncorrect = true;
      }
    }
    
    function cancel() {
      $window.location.href = '/ng/project';
    }
    
  }
  
})();