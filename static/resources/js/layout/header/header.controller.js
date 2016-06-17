(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.header')
    .controller('HeaderController', HeaderController);
  
  HeaderController.$inject = ['$scope', '$window', 'getParameterByName'];
  
  function HeaderController($scope, $window, getParameterByName) {
    var vm = this;
    if($window.location.search) {
      vm.searchInput = getParameterByName('q', $window.location.search);
      console.log('vm.searchInput at header:' + vm.searchInput);
    }
  }
  
})();