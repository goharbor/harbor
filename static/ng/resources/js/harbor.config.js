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
      .factory('getParameterByName', getParameterByName)
      .filter('dateL', localizeDate)
      .filter('tr', tr);
  
  function getParameterByName() {
    return get;
    function get(name, url) {
      name = name.replace(/[\[\]]/g, "\\$&");
      var regex = new RegExp("[?&]" + name + "(=([^&#]*)|&|#|$)"),
          results = regex.exec(url);
      if (!results) return null;
      if (!results[2]) return '';
      return decodeURIComponent(results[2].replace(/\+/g, " "));
    }
  }
  
  function localizeDate() {
    return filter;
    
    function filter(input, pattern) {
      return moment(new Date(input || '')).format(pattern);
    }
  }
  
  tr.$inject = ['I18nService'];
  
  function tr(I18nService) {
    return tr;
    function tr(label, params) {
      var currentLanguage = I18nService().getCurrentLanguage();
      var result = '';
      if(label && label.length > 0){
        result = I18nService().getValue(label, currentLanguage); 
      }
      if(angular.isArray(params)) {
        angular.forEach(params, function(value, index) {
          result = result.replace('$' + index, params[index]);
        });
      }
      return result;
    }
  }  
    
})();