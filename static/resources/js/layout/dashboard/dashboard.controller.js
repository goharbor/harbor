(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.dashboard')
    .controller('DashboardController', DashboardController);
  
  DashboardController.$inject = ['$scope'];
  
  function DashboardController($scope) {
    var vm = this;
    vm.customBodyHeight = {'height': '165px'};
    
    //Error message dialog handler for dashboard.
    $scope.$on('modalTitle', function(e, val) {
      vm.modalTitle = val;
    });
    
    $scope.$on('modalMessage', function(e, val) {
      vm.modalMessage = val;
    });
       
    $scope.$on('raiseError', function(e, val) {
      if(val) {   
        vm.action = function() {
          $scope.$broadcast('showDialog', false);
        };
        vm.contentType = 'text/plain';
        vm.confirmOnly = true;      
        $scope.$broadcast('showDialog', true);
      }
    });
  }
  
})();