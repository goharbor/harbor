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
    .module('harbor.app')
    .config(function($interpolateProvider){
      $interpolateProvider.startSymbol('//');
      $interpolateProvider.endSymbol('//');
    })
    .config(function($httpProvider) { 
      //initialize get if not there
      if (!$httpProvider.defaults.headers.get) {
          $httpProvider.defaults.headers.get = {};    
      }    
  
      // Answer edited to include suggestions from comments
      // because previous version of code introduced browser-related errors
  
      //disable IE ajax request caching
      $httpProvider.defaults.headers.get['If-Modified-Since'] = 'Mon, 26 Jul 1997 05:00:00 GMT';
      // extra
      $httpProvider.defaults.headers.get['Cache-Control'] = 'no-cache';
      $httpProvider.defaults.headers.get['Pragma'] = 'no-cache';
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
          /^\/login$/,
          /^\/api\/targets\/ping$/,
          /^\/api\/users\/current$/,
          /^\/api\/repositories$/,
          /^\/api\/projects\/[0-9]+\/members\/current$/
        ];
        var isExcluded = false;
        for(var i in exclusion) {
          isExcluded = exclusion[i].test(url);
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
      if(d.getTime() <= 0) {return '-';}
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
