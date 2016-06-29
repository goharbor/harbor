(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.dashboard')
    .controller('DashboardController', DashboardController);
  
  DashboardController.$inject = ['$scope'];
  
  function DashboardController($scope) {
    var vm = this;
    vm.customBodyHeight = {'height': '165px'};
  }
  
})();