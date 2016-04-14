(function() {

  'use strict';
  
  angular
    .module('harbor.log')
    .controller('ListLogController', ListLogController);
    
  ListLogController.$inject  = ['ListLogService']
  
  function ListLogController(ListLogService) {
    
  }
  
})();