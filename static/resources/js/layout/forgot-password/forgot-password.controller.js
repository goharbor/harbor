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
    vm.toggleInProgress = false;
    
    function reset(){
      vm.hasError = false;
      vm.errorMessage = '';
    }
    
    function sendMail(user) {
      if(user && angular.isDefined(user.email)) { 
        vm.toggleInProgress = true;
        SendMailService(user.email)
          .success(sendMailSuccess)
          .error(sendMailFailed);
      }
    }
    
    function sendMailSuccess(data, status) {
      vm.toggleInProgress = false;
      $scope.$broadcast('showDialog', true);
    }
    
    function sendMailFailed(data) {
      vm.toggleInProgress = false;
      vm.hasError = true;
      vm.errorMessage = data;
      console.log('Failed send mail:' + data);
    }
    
    function confirm() {
      $window.location.href = '/';
    }
   
    
  }
  
})();