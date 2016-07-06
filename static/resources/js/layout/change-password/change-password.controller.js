(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.change.password')
    .controller('ChangePasswordController', ChangePasswordController);
  
  ChangePasswordController.$inject = ['ChangePasswordService', 'UpdateUserService', '$filter', 'trFilter', '$scope', '$window', 'currentUser'];
  
  function ChangePasswordController(ChangePasswordService, UpdateUserService, $filter, trFilter, $scope, $window, currentUser) {

    var vm = this;
    vm.isOpen = false;
 
    vm.hasError = false;
    vm.errorMessage = '';
    
    vm.reset = reset;    
    
    vm.confirm = confirm;
    vm.updatePassword = updatePassword;
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
                
    function updatePassword(user) {
      if(user && angular.isDefined(user.oldPassword) && angular.isDefined(user.password)) {
        vm.action = vm.confirm;
        ChangePasswordService(userId, user.oldPassword, user.password)
          .success(changePasswordSuccess)
          .error(changePasswordFailed);
      }
     
    }
    
    function changePasswordSuccess(data, status) {
      vm.modalTitle = $filter('tr')('change_password', []);
      vm.modalMessage = $filter('tr')('successful_changed_password', []);
      $scope.$broadcast('showDialog', true);
    }
    
    function changePasswordFailed(data, status) {

      var message;
      $scope.$emit('modalTitle', $filter('tr')('error'));
      console.log('Failed to change password:' + data);
      if(data == 'old_password_is_not_correct') {
        message = $filter('tr')('old_password_is_incorrect');
      }else{
        message = $filter('tr')('failed_to_change_password');
      }

      $scope.$emit('modalMessage', message);
      $scope.$emit('raiseError', true);
    }
    
    function cancel(form) {
      $window.location.href = '/dashboard';
    }
    
  }
  
})();
