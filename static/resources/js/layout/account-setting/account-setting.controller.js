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
    vm.confirm = confirm;
    vm.updateUser = updateUser;
    vm.cancel = cancel;
    
    $scope.user = currentUser.get();
    var userId = $scope.user.user_id;

    //Error message dialog handler for account setting.
    $scope.$on('modalTitle', function(e, val) {
      vm.modalTitle = val;
    });
    
    $scope.$on('modalMessage', function(e, val) {
      vm.modalMessage = val;
    });
       
    $scope.$on('raiseError', function(e, val) {
      if(val) {   
        vm.action = function() {
          $scope.$broadcast('showDialog', false);
        };
        vm.contentType = 'text/plain';    
        vm.confirmOnly = true;  
        $scope.$broadcast('showDialog', true);
      }
    });
        
    function reset() {
      $scope.form.$setUntouched();
      $scope.form.$setPristine();
      vm.hasError = false;
      vm.errorMessage = '';
    }
     
    function confirm() {     
      $window.location.href = '/dashboard';
    }    
                
    function updateUser(user) {
      vm.confirmOnly = true;
      vm.action = vm.confirm;
      if(user && angular.isDefined(user.username) && angular.isDefined(user.password) && 
            angular.isDefined(user.realname)) {
        UpdateUserService(userId, user)
          .success(updateUserSuccess)
          .error(updateUserFailed); 
        currentUser.set(user);        
      }
    }
        
    function updateUserSuccess(data, status) {
      vm.modalTitle = $filter('tr')('change_profile', []);
      vm.modalMessage = $filter('tr')('successful_changed_profile', []);
      $scope.$broadcast('showDialog', true);
    }
    
    function updateUserFailed(data, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      var message;
      if(status === 409) {
        message = $filter('tr')('email_has_been_taken');
      }else{
        message = $filter('tr')('failed_to_update_user') + data;
      }
      $scope.$emit('modalMessage', message);
      $scope.$emit('raiseError', true);
      console.log('Failed to update user.');
    }
    
    function cancel(form) {
      $window.location.href = '/dashboard';
    }
    
  }
  
})();