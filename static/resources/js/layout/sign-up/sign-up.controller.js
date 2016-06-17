(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.sign.up')
    .controller('SignUpController', SignUpController);
 
  SignUpController.$inject = ['SignUpService', '$window'];
  
  function SignUpController(SignUpService, $window) {
    var vm = this;
     
    vm.user = {};
    vm.signUp = signUp;
    
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
      alert('Signed up successfully');
      $window.location.href = '/';
    }
    
    function signUpFailed(data, status) {
      console.log('Signed up failed.');
    }
    
  }
  
})();