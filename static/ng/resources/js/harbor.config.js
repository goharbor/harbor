(function() {
    'use strict';
    angular
      .module('harbor.app')
      .config(function($interpolateProvider){
        $interpolateProvider.startSymbol('//');
        $interpolateProvider.endSymbol('//');
      })
      .config(function($httpProvider) {
        $httpProvider.defaults.headers.common = {'Accept': 'application/json, text/javascript, */*; q=0.01'};     
      })
      .filter('dateL', localizeDate);
    
  function localizeDate() {
    return filter;
    
    function filter(input, pattern) {
      return moment(new Date(input || '')).format(pattern);
    }
  }  
    
})();