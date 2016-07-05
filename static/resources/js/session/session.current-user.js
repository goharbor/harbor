(function() { 

  'use strict';
  
  angular
    .module('harbor.session')
    .controller('CurrentUserController', CurrentUserController);
 
  CurrentUserController.$inject = ['$scope', 'CurrentUserService', 'currentUser', '$window', '$document'];
  
  function CurrentUserController($scope, CurrentUserService, currentUser, $window, $document) {
    
    var vm = this;
         
    CurrentUserService()
      .then(getCurrentUserComplete)
      .catch(getCurrentUserFailed);
        
    function getCurrentUserComplete(response) {
      if(angular.isDefined(response)) {
        currentUser.set(response.data);  
        if(location.pathname === '/') {
          $window.location.href = '/dashboard';
        }
      }   
    }
    
    function getCurrentUserFailed(e){
      console.log('No session of current user.');
    }   
  }
 
})();