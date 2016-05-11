(function() { 

  'use strict';
  
  angular
    .module('harbor.session')
    .controller('CurrentUserController', CurrentUserController)
    
  CurrentUserController.$inject = ['CurrentUserService', 'currentUser', '$scope', '$timeout'];
  
  function CurrentUserController(CurrentUserService, currentUser, $scope, $timeout) {
    
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
      console.log('Have not logged in yet.');
    }
  }

})();