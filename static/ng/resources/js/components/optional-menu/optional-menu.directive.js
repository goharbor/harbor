(function() {
  
  'use strict';
  
  angular
    .module('harbor.optional.menu')
    .directive('optionalMenu', optionalMenu);

  OptionalMenuController.$inject = ['$window', 'I18nService', 'LogOutService', 'currentUser', '$timeout'];

  function OptionalMenuController($window, I18nService, LogOutService, currentUser, $timeout) {
    var vm = this;
    
    vm.currentLanguage = I18nService().getCurrentLanguage();
    vm.languageName = I18nService().getLanguageName(vm.currentLanguage);
    console.log('current language:' + I18nService().getCurrentLanguage());

    vm.isLoggedIn = false;    
    $timeout(function() {
      vm.user = currentUser.get();
      console.log('current user:' + angular.toJson(vm.user));
      if(angular.isDefined(vm.user)) {
        vm.isLoggedIn = true;
      }
    }, 300);

    vm.setLanguage = setLanguage;     
    vm.logOut = logOut;
    
    function setLanguage(name) {
      I18nService().setCurrentLanguage(name);
      $window.location.reload();
    }
        
    function logOut() {
      LogOutService()
        .success(logOutSuccess)
        .error(logOutFailed);
    }
    function logOutSuccess(data, status) {
      currentUser.unset();
      $window.location.href= '/ng';
    }
    function logOutFailed(data, status) {
      console.log('Failed to log out:' + data);
    }
  }
  
  function optionalMenu() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/optional-menu/optional-menu.directive.html',
      'scope': true,
      'controller': OptionalMenuController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();