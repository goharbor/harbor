(function() { 

  'use strict';
  
  angular
    .module('harbor.session')
    .controller('CurrentUserController', CurrentUserController);
 
  CurrentUserController.$inject = ['CurrentUserService', 'currentUser', '$scope', '$timeout', '$window'];
  
  function CurrentUserController(CurrentUserService, currentUser, $scope, $timeout, $window) {
    
    var vm = this;
    
    CurrentUserService()
      .then(getCurrentUserComplete)
      .catch(getCurrentUserFailed);
      
    function getCurrentUserComplete(response) {
      console.log('Successful logged in.');
      $timeout(function(){
        $scope.$broadcast('currentUser', response.data);
        currentUser.set(response.data);
      }, 50);
    }
    
    function getCurrentUserFailed(e){
      var url = location.pathname;
      var exclusions = ['/ng', '/ng/forgot_password', '/ng/sign_up', '/ng/reset_password'];
      for(var i = 0; i < exclusions.length; i++) {
        if(exclusions[i]===url) {
          return;
        }
      }     
      $window.location.href = '/ng';
    }
  }
 
})();