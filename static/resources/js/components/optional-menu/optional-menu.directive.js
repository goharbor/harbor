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
    
    I18nService().setCurrentLanguage(vm.currentLanguage);        
    
    console.log('current language:' + vm.languageName);

    vm.supportLanguages = I18nService().getSupportLanguages(); 
    vm.user = currentUser.get();
    vm.setLanguage = setLanguage;     
    vm.logOut = logOut;
        
    function setLanguage(language) {
      I18nService().setCurrentLanguage(language);
      var hash = $window.location.hash;
      $window.location.href = '/language?lang=' + language + '&hash=' + encodeURIComponent(hash);    
    }
    function logOut() {
      LogOutService()
        .success(logOutSuccess)
        .error(logOutFailed);
    }
    function logOutSuccess(data, status) {
      currentUser.unset();
      $window.location.href= '/';
    }
    function logOutFailed(data, status) {
      console.log('Failed to log out:' + data);
    }
  }
  
  function optionalMenu() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/optional_menu?timestamp=' + new Date().getTime(),
      'scope': true,
      'controller': OptionalMenuController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();