/*
    Copyright (c) 2016 VMware, Inc. All Rights Reserved.
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
        
        http://www.apache.org/licenses/LICENSE-2.0
        
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/
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