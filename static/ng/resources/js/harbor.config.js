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
      .filter('dateL', localizeDate)
      .filter('tr', tr);
    
  function localizeDate() {
    return filter;
    
    function filter(input, pattern) {
      return moment(new Date(input || '')).format(pattern);
    }
  }
  
  tr.$inject = ['I18nService'];
  
  function tr(I18nService) {
    return tr;
    function tr(label) {
      var currentLanguage = I18nService().getCurrentLanguage();
      if(label && label.length > 0){
        return I18nService().getValue(label, currentLanguage); 
      }
      return '';
    }
  }  
    
})();