(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.reset.password')
    .controller('ResetPasswordController', ResetPasswordController);
  
  ResetPasswordController.$inject = ['$location', 'ResetPasswordService', '$window', 'getParameterByName'];
  
  function ResetPasswordController($location, ResetPasswordService, $window, getParameterByName) {
    var vm = this;
    vm.resetUuid = getParameterByName('reset_uuid', $location.absUrl());
    
    vm.reset = reset;
    vm.resetPassword = resetPassword;
    vm.cancel = cancel;
    
    vm.hasError = false;
    vm.errorMessage = '';
    
    function reset() {
      vm.hasError = false;
      vm.errorMessage = '';
    }    
        
    function resetPassword(user) {
      if(user && angular.isDefined(user.password)) {
        console.log('rececived password:' + user.password + ', reset_uuid:' + vm.resetUuid);
        ResetPasswordService(vm.resetUuid, user.password)
          .success(resetPasswordSuccess)
          .error(resetPasswordFailed);
      }
    }
    
    function resetPasswordSuccess(data, status) {
      $window.location.href = '/';
    }
    
    function resetPasswordFailed(data) {
      vm.hasError = true;
      vm.errorMessage = data;
      console.log('Failed reset password:' + data);
    }
    
    function cancel(form) {
      if(form) {
        form.$setPristine();
      }
    }
  }
  
})();