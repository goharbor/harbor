(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.account.setting')
    .controller('AccountSettingController', AccountSettingController);
  
  AccountSettingController.$inject = ['ChangePasswordService', '$scope', '$window', 'currentUser'];
  
  function AccountSettingController(ChangePasswordService, $scope, $window, currentUser) {
    var vm = this;
    vm.isOpen = false;
    vm.user = {};
    
    vm.hasError = false;
    vm.errorMessage = '';
    
    vm.reset = reset;    
    vm.toggleChangePassword = toggleChangePassword;
    vm.changeProfile = changeProfile;
    vm.changePassword= changePassword;
    vm.cancel = cancel;
    
   
    vm.user = currentUser.get();
    
    function reset() {
      vm.hasError = false;
      vm.errorMessage = '';
    }
     
    function toggleChangePassword() {
      if(vm.isOpen) {
        vm.isOpen = false;
      }else{
        vm.isOpen = true;
      }     
    }
        
    function getCurrentUserFailed(data) {
      console.log('Failed get current user:' + data);
    }
    
    function changeProfile(user) {
      console.log(user);
    }
    
    function changePassword(user) {
      if(user && angular.isDefined(user.oldPassword) && angular.isDefined(user.password)) {
        ChangePasswordService(vm.user.user_id, user.oldPassword, user.password)
          .success(changePasswordSuccess)
          .error(changePasswordFailed);
      }
    }
    
    function changePasswordSuccess(data, status) {
      $window.location.href = '/project';
    }
    
    function changePasswordFailed(data, status) {
      console.log('Failed changed password:' + data);
      if(data == 'old_password_is_not_correct') {
        vm.hasError = true;
        vm.errorMessage = 'old_password_is_incorrect';
      }
    }
    
    function cancel(form) {
      if(form) {
        form.$setPristine();
      }
      $window.location.href = '/project';
    }
    
  }
  
})();