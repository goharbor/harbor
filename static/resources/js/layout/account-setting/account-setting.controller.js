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
    vm.confirm = confirm;
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
    
    function confirm() {     
      $window.location.href = '/dashboard';
    }    
                
    function updateUser(user) {
      if(vm.isOpen){
        if(user && angular.isDefined(user.oldPassword) && angular.isDefined(user.password)) {
          ChangePasswordService(userId, user.oldPassword, user.password)
            .success(changePasswordSuccess)
            .error(changePasswordFailed);
        }
      }else{
        if(user && angular.isDefined(user.username) && angular.isDefined(user.password) && 
            angular.isDefined(user.realname)) {
          UpdateUserService(userId, user)
            .success(updateUserSuccess)
            .error(updateUserFailed); 
          currentUser.set(user);        
        }
      }
    }
    
    function changePasswordSuccess(data, status) {
      vm.modalTitle = $filter('tr')('change_password', []);
      vm.modalMessage = $filter('tr')('successful_changed_password', []);
      $scope.$broadcast('showDialog', true);
    }
    
    function changePasswordFailed(data, status) {
      console.log('Failed changed password:' + data);
      if(data == 'old_password_is_not_correct') {
        vm.hasError = true;
        vm.errorMessage = 'old_password_is_incorrect';
      }
    }
    
    function updateUserSuccess(data, status) {
      vm.modalTitle = $filter('tr')('change_profile', []);
      vm.modalMessage = $filter('tr')('successful_changed_profile', []);
      $scope.$broadcast('showDialog', true);
    }
    
    function updateUserFailed(data, status) {
      console.log('Failed update user.');
    }
    
    function cancel(form) {
      $window.location.href = '/dashboard';
    }
    
  }
  
})();