(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.forgot.password')
    .controller('ForgotPasswordController', ForgotPasswordController);
  
  ForgotPasswordController.$inject = ['SendMailService', '$window', '$scope'];
  
  function ForgotPasswordController(SendMailService, $window, $scope) {
    var vm = this;
    
    vm.hasError = false;
    vm.show = false;
    vm.errorMessage = '';
    
    vm.reset = reset;
    vm.sendMail = sendMail;
    
    vm.confirm = confirm;
    
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
      $scope.$broadcast('showDialog', true);
    }
    
    function sendMailFailed(data) {
      vm.hasError = true;
      vm.errorMessage = data;
      console.log('Failed send mail:' + data);
    }
    
    function confirm() {
      $window.location.href = '/';
    }
   
    
  }
  
})();