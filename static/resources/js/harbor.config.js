(function() {
  'use strict';
  angular
    .module('harbor.app')
    .config(function($interpolateProvider){
      $interpolateProvider.startSymbol('//');
      $interpolateProvider.endSymbol('//');
    })
    .config(function($httpProvider) {
//      if(!$httpProvider.defaults.headers) {
//        $httpProvider.defaults.headers = {};
//      }
//      $httpProvider.defaults.headers.get['If-Modified-Since'] = 'Mon, 26 Jul 1997 05:00:00 GMT';
//      $httpProvider.defaults.headers.get['Cache-Control'] = 'no-cache';
//      $httpProvider.defaults.headers.get['Pragma'] = 'no-cache';
      
      $httpProvider.defaults.headers.common = {'Accept': 'application/json, text/javascript, */*; q=0.01'};     
      $httpProvider.interceptors.push('redirectInterceptor');
    })
    .service('redirectInterceptor', RedirectInterceptorService)
    .factory('getParameterByName', getParameterByName)
    .filter('dateL', localizeDate)
    .filter('tr', tr);
   
  RedirectInterceptorService.$inject = ['$q', '$window', '$location'];
  
  function RedirectInterceptorService($q, $window, $location) {
    return {
      'responseError': function(rejection) {
        var url = rejection.config.url;
        console.log('url:' + url);
        var exclusion = [
         '/',
         '/search',
         '/reset_password',
         '/sign_up', 
         '/forgot_password', 
         '/api/targets/ping',
         '/api/users/current',
         '/api/repositories',
          /^\/api\/projects\/[0-9]+\/members\/current$/
        ];
        var isExcluded = false;
        for(var i in exclusion) {
          switch(typeof(exclusion[i])) {
          case 'string':
            isExcluded = (exclusion[i] === url);
            break;
          case 'object':
            isExcluded = exclusion[i].test(url);
            break;
          }
          if(isExcluded) {
            break;
          }
        }        
        if(!isExcluded && rejection.status === 401) {
          $window.location.href = '/?last_url=' + encodeURIComponent(location.pathname + '#' + $location.url());
          return;
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