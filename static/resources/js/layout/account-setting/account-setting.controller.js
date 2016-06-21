(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.account.setting')
    .controller('AccountSettingController', AccountSettingController);
  
  AccountSettingController.$inject = ['ChangePasswordService', 'UpdateUserService', '$filter', 'trFilter', '$scope', '$window', 'currentUser'];
  
  function AccountSettingController(ChangePasswordService, UpdateUserService, $filter, trFilter, $scope, $window, currentUser) {
    var vm = this;
    vm.isOpen = false;
 
    vm.hasError = false;
    vm.errorMessage = '';
    
    vm.reset = reset;    
    vm.toggleChangePassword = toggleChangePassword;
    vm.confirmToUpdate = confirmToUpdate;
    vm.updateUser = updateUser;
    vm.cancel = cancel;
    
    $scope.user = currentUser.get();
    var userId = $scope.user.user_id;
        
    function reset() {
      $scope.form.$setUntouched();
      $scope.form.$setPristine();
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
    
    function confirmToUpdate(user) {     
      vm.user = user;
      if(vm.isOpen) {
        if(vm.user && angular.isDefined(user.oldPassword) && angular.isDefined(user.password)) {
          vm.modalTitle = $filter('tr')('change_password', []);
          vm.modalMessage = $filte('tr')('confirm_to_change_password', []);
          return true;
        }
      }else{
        if(vm.user && angular.isDefined(vm.user.username) && angular.isDefined(vm.user.password) && 
            angular.isDefined(vm.user.realname)) {
          vm.modalTitle = $filter('tr')('change_profile', []);
          vm.modalMessage = $filter('tr')('confirm_to_change_profile', []);
          return true;
        }
      }
      
      vm.modalTitle = $filter('tr')('form_is_invalid');
      vm.modalMessage = $filter('tr')('form_is_invalid_message', []);
      return false;
    }    
                
    function updateUser() {
      if(vm.isOpen){
        ChangePasswordService(userId, vm.user.oldPassword, vm.user.password)
          .success(changePasswordSuccess)
          .error(changePasswordFailed);
      }else{
        UpdateUserService(userId, vm.user)
          .success(updateUserSuccess)
          .error(updateUserFailed); 
        currentUser.set(vm.user);        
      }
    }
    
    function changePasswordSuccess(data, status) {
      $window.location.href = '/dashboard';
    }
    
    function changePasswordFailed(data, status) {
      console.log('Failed changed password:' + data);
      if(data == 'old_password_is_not_correct') {
        vm.hasError = true;
        vm.errorMessage = 'old_password_is_incorrect';
      }
    }
    
    function updateUserSuccess(data, status) {
      $window.location.href = '/dashboard';
    }
    
    function updateUserFailed(data, status) {
      console.log('Failed update user.');
    }
    
    function cancel(form) {
      $window.location.href = '/dashboard';
    }
    
  }
  
})();