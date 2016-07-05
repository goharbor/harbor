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
      $httpProvider.interceptors.push('redirectInterceptor');
    })
    .service('redirectInterceptor', RedirectInterceptorService)
    .factory('getParameterByName', getParameterByName)
    .filter('dateL', localizeDate)
    .filter('tr', tr);
   
  RedirectInterceptorService.$inject = ['$q', '$window'];
  
  function RedirectInterceptorService($q, $window) {
    return {
      'responseError': function(rejection) {
        var pathname = $window.location.pathname;
        var exclusions = ['/', '/search', '/reset_password', '/sign_up', '/forgot_password'];
        var isExcluded = false;
        for(var i in exclusions) {
          if(exclusions[i] === pathname) {
            isExcluded = true;
            break;
          }
        }
        if(rejection.status === 401 && !isExcluded) {
          $window.location.href = '/';
        }
        return $q.reject(rejection);
      }
    };
  }
  
  function getParameterByName() {
    return get;
    function get(name, url) {
      name = name.replace(/[\[\]]/g, "\\$&");
      var regex = new RegExp("[?&]" + name + "(=([^&#]*)|&|#|$)"),
          results = regex.exec(url);
      if (!results) {
        return null;
      }
      
      if (!results[2]) {
        return '';
      }
      
      return decodeURIComponent(results[2].replace(/\+/g, " "));
    }
  }
  
  function localizeDate() {
    return filter;
    
    function filter(input, pattern) {
      var d = new Date(input || '');
      if(d.getTime() <= 0) return '-';
      return moment(d).format(pattern);
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