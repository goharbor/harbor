(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.sign.up')
    .controller('SignUpController', SignUpController);
 
  SignUpController.$inject = ['$scope', 'SignUpService', '$window', '$filter', 'trFilter'];
  
  function SignUpController($scope, SignUpService, $window, $filter, trFilter) {
    var vm = this;
     
    vm.user = {};
    vm.signUp = signUp;
    vm.confirm = confirm;
        
    //Error message dialog handler for signing up.
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
    
    function signUp(user) {
      var userObject = {
        'username': user.username,
        'email': user.email,
        'password': user.password,
        'realname': user.fullName,
        'comment': user.comment
      };
      
      vm.action = vm.confirm;
      
      SignUpService(userObject)
        .success(signUpSuccess)
        .error(signUpFailed);        
    } 
   
    function signUpSuccess(data, status) {
      var title;
      var message;
      if(vm.targetType) {
        title = $filter('tr')('add_new_title');
        message = $filter('tr')('successful_added');
      }else{
        title = $filter('tr')('sign_up');
        message = $filter('tr')('successful_signed_up');
      }
      vm.modalTitle = title;
      vm.modalMessage = message;
      $scope.$broadcast('showDialog', true);
    }
    
    function signUpFailed(data, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      var message;
      if(vm.targetType) {
        message = $filter('tr')('failed_add_user');
      }else{
        message = $filter('tr')('failed_sign_up');
      }
      $scope.$emit('modalMessage', message);
      $scope.$emit('raiseError', true);
      
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