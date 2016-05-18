(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.system.management')
    .controller('SystemManagementController', SystemManagementController);
    
  function SystemManagementController() {
    var vm = this;
    vm.registrationOptions = [
      { 
        'name': 'on',
        'value': true   
      },
      {
        'name': 'off',
        'value': false
      }
    ];
    vm.currentRegistration = { 
      'name': 'on',
      'value': true   
    };
    
    vm.changeSettings = changeSettings;
    
    vm.selectRegistration = selectRegistration;
    
    function selectRegistration() {
      
    }
    
    function changeSettings(system) {
      console.log(system);
    }
  }
  
})();