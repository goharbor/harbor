(function() {
  
  'use strict';
  
  angular
    .module('harbor.system.management')
    .directive('configuration', configuration);
  
  ConfigurationController.$inject = [];
  
  function ConfigurationController() {
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
  
  function configuration() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/system-management/configuration.directive.html',
      'scope': true,
      'controller': ConfigurationController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
  
})();