(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.admin.option')
    .controller('AdminOptionController', AdminOptionController);
  
  function AdminOptionController() {
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