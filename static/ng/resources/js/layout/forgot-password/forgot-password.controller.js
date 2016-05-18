(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.forgot.password')
    .controller('ForgotPasswordController', ForgotPasswordController);
  
  ForgotPasswordController.$inject = ['SendMailService', '$window'];
  
  function ForgotPasswordController(SendMailService, $window) {
    var vm = this;
    
    vm.hasError = false;
    vm.errorMessage = '';
    
    vm.reset = reset;
    vm.sendMail = sendMail;
    
    function reset(){
      vm.hasError = false;
      vm.errorMessage = '';
    }
    
    function sendMail(user) {
      if(user && angular.isDefined(user.email)) { 
        SendMailService(user.email)
          .success(sendMailSuccess)
          .error(sendMailFailed);
      }
    }
    
    function sendMailSuccess(data, status) {
      $window.location.href = '/ng';
    }
    
    function sendMailFailed(data) {
      vm.hasError = true;
      vm.errorMessage = data;
      console.log('Failed send mail:' + data);
    }
  }
  
})();