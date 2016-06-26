(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.admin.option')
    .controller('AdminOptionController', AdminOptionController);
  
  AdminOptionController.$inject = ['$scope'];
  
  function AdminOptionController($scope) {
    
    $scope.subsSubPane = 296;   
    
    var vm = this;
    vm.toggle = false;
    vm.toggleAdminOption = toggleAdminOption;
    
    function toggleAdminOption() {
      if(vm.toggle) {
        vm.toggle = false;
      }else{
        vm.toggle = true;
      }
    }
  }
  
})();