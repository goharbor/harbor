(function() { 

  'use strict';
  
  angular
    .module('harbor.session')
    .controller('CurrentUserController', CurrentUserController)
    
  CurrentUserController.$inject = ['CurrentUserService', '$scope', '$timeout', '$window'];
  
  function CurrentUserController(CurrentUserService, $scope, $timeout, $window) {
    
    var vm = this;
    
    CurrentUserService()
      .then(getCurrentUserComplete)
      .catch(getCurrentUserFailed);
      
    function getCurrentUserComplete(response) {
      console.log('Successful logged in.');
      $timeout(function(){
        $scope.$broadcast('currentUser', response.data);
      }, 50);
    }
    
    function getCurrentUserFailed(e){
      console.log('Have not logged in yet.');
      $timeout(function(){
        $scope.$broadcast('currentUser', null);
      });
    }
  }

})();