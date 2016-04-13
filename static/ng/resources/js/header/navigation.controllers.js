(function() {

  'use strict';
  
  angular
    .module('harbor.header')
    .controller('NavigationController', NavigationController);
  
  NavigationController.$inject = ['navigationTabs'];
  
  
  function NavigationController(navigationTabs) {
    var vm = this;
    vm.tabs = navigationTabs();
  }
  
})();