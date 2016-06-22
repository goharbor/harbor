(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.sign.up')
    .controller('SignUpController', SignUpController);
 
  SignUpController.$inject = ['$scope', 'SignUpService', '$window'];
  
  function SignUpController($scope, SignUpService, $window) {
    var vm = this;
     
    vm.user = {};
    vm.signUp = signUp;
    vm.confirm = confirm;
    
    function signUp(user) {
      var userObject = {
        'username': user.username,
        'email': user.email,
        'password': user.password,
        'realname': user.fullName,
        'comment': user.comments
      };
      SignUpService(userObject)
        .success(signUpSuccess)
        .error(signUpFailed);        
    }
    
   
    function signUpSuccess(data, status) {
      console.log('Signed up successfully:' + status);
      $scope.$broadcast('showDialog', true);
    }
    
    function signUpFailed(data, status) {
      console.log('Signed up failed.');
    }
    
    function confirm() {
      if(location.pathname === '/add_new') {
        $window.location.href = '/dashboard';  
      }else{
        $window.location.href = '/';  
      }
      
    }
    
  }
  
})();