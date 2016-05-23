(function() { 

  'use strict';
  
  angular
    .module('harbor.session')
    .controller('CurrentUserController', CurrentUserController);
 
  CurrentUserController.$inject = ['CurrentUserService', 'currentUser', '$window'];
  
  function CurrentUserController(CurrentUserService, currentUser, $window) {
    
    var vm = this;
    
    CurrentUserService()
      .then(getCurrentUserComplete)
      .catch(getCurrentUserFailed);
        
    function getCurrentUserComplete(response) {
      if(angular.isDefined(response)) {
        currentUser.set(response.data);  
      }   
    }
    
    function getCurrentUserFailed(e){
      var url = location.pathname;
      var exclusions = [
        '/ng',
        '/ng/forgot_password', 
        '/ng/sign_up', 
        '/ng/reset_password',
        '/ng/search'
      ];
      for(var i = 0; i < exclusions.length; i++) {
        if(exclusions[i]===url) {
          return;
        }
      }     
      $window.location.href = '/ng';
    }   
  }
 
})();