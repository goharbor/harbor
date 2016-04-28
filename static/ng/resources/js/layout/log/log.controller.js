(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.log')
    .controller('LogController', LogController);
    
  LogController.$inject = ['$scope'];
    
  function LogController($scope) {
    var vm = this;
    $scope.$on('currentProjectId', function(e, val) {
      console.log('received currentProjecjtId: ' + val + ' in LogController');
      vm.projectId = val;
    });
  }
  
})();