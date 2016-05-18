(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.header')
    .controller('HeaderController', HeaderController);
  
  HeaderController.$inject = ['$scope'];
  
  function HeaderController($scope) {
    var vm = this;
  }
  
})();