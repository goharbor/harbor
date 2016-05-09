(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.header')
    .controller('HeaderController', HeaderController);
  
  HeaderController.$inject = ['CurrentUserService', '$scope'];
  
  function HeaderController(CurrentUserService, $scope) {
    var vm = this;
    
    vm.isLoggedIn = true;
    vm.currentUser = {};

    CurrentUserService()
      .then(currentUserSucess)
      .catch(currentUserFailed);      
    
    function currentUserSucess(response) {
      vm.isLoggedIn = true;
      vm.currentUser.username = response.data.username;
      console.log('vm.currentUser.username:' + vm.currentUser.username);
    }
    
    function currentUserFailed(e) {
//      vm.isLoggedIn = false;
    }
  }
  
})();