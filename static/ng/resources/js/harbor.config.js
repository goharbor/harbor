(function() {
   'use strict';
   angular
     .module('harbor.app')
     .config(function($interpolateProvider){
        $interpolateProvider.startSymbol('//');
        $interpolateProvider.endSymbol('//');
      });
    
})();