(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.reset.password')
    .controller('ResetPasswordController', ResetPasswordController);
  
  ResetPasswordController.$inject = ['$location', 'ResetPasswordService', '$window'];
  
  function getParameterByName(name, url) {
    name = name.replace(/[\[\]]/g, "\\$&");
    var regex = new RegExp("[?&]" + name + "(=([^&#]*)|&|#|$)"),
        results = regex.exec(url);
    if (!results) return null;
    if (!results[2]) return '';
    return decodeURIComponent(results[2].replace(/\+/g, " "));
  }
  
  function ResetPasswordController($location, ResetPasswordService, $window) {
    var vm = this;
    vm.resetUuid = getParameterByName('reset_uuid', $location.absUrl());
    vm.resetPassword = resetPassword;
    console.log(vm.resetUuid);
    function resetPassword(user) {
      console.log('rececived password:' + user.password + ', reset_uuid:' + vm.resetUuid);
      ResetPasswordService(vm.resetUuid, user.password)
        .success(resetPasswordSuccess)
        .error(resetPasswordFailed);
    }
    
    function resetPasswordSuccess(data, status) {
      console.log('Successful reset password:' + data);
      $window.location.href = '/ng';
    }
    
    function resetPasswordFailed(data) {
      console.log('Failed reset password:' + data);
    }
    
    
    
  }
  
})();