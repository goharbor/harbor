(function() {
  'use strict';
  
  angular
    .module('harbor.app')
    .run(CurrentUser);
  
  CurrentUser.$inject = ['CurrentUserService', '$log'];
  
  function CurrentUser(CurrentUserService, $log) {
    
    CurrentUserService()
      .then(getCurrentUserComplete)
      .catch(getCurrentUserFailed);
      
      function getCurrentUserComplete(data) {
        $log.info(data.data);
      }
      
      function getCurrentUserFailed(e){
        $log.info(e);
      }
      
  }
})();