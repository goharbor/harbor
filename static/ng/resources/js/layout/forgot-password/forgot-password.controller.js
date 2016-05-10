(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.forgot.password')
    .controller('ForgotPasswordController', ForgotPasswordController);
  
  ForgotPasswordController.$inject = ['SendMailService'];
  
  function ForgotPasswordController(SendMailService) {
    var vm = this;
    vm.hasError = false;
    vm.errorMessage = '';
    vm.sendMail = sendMail;
    
    function sendMail(user) {
      vm.hasError = false;
      console.log('Email address:' + user.email);
      SendMailService(user.email)
        .success(sendMailSuccess)
        .error(sendMailFailed);
    }
    
    function sendMailSuccess(data, status) {
      console.log('Successful send mail:' + data);
    }
    
    function sendMailFailed(data) {
      vm.hasError = true;
      vm.errorMessage = data;
      console.log('Failed send mail:' + data);
    }
  }
  
})();