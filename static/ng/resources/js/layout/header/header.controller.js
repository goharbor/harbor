(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.header')
    .controller('HeaderController', HeaderController);
  
  HeaderController.$inject = ['$scope', 'I18nService', '$cookies', '$window'];
  
  function HeaderController($scope, I18nService, $cookies, $window) {
    
  }
  
})();