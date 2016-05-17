(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.dashboard', [
      'harbor.services.project',
      'harbor.services.repository',
      'harbor.services.log'
    ]);
  
})();