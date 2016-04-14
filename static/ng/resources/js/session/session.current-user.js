(function() { 

  'use strict';
  
  angular
    .module('harbor.session')
    .controller('CurrentUserController', CurrentUserController)
    .directive('currentUser', currentUser);
  
  CurrentUserController.$inject = ['CurrentUserService', '$log', '$window'];
  
  function CurrentUserController(CurrentUserService, $log, $window) {
    
    CurrentUserService()
      .then(getCurrentUserComplete)
      .catch(getCurrentUserFailed);
      
    function getCurrentUserComplete(data) {
      $log.info('login success');
    }
    
    function getCurrentUserFailed(e){
      if(e.status == 401) {
        $window.location = '/ng';
      }
    }
  }

  function currentUser() {
    var directive = {
      restrict: 'A',
      controller: CurrentUserController,
      bindToController: true
    }
    return directive;
  }

})();