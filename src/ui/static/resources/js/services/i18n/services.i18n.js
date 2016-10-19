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
    .module('harbor.services.i18n')
    .factory('I18nService', I18nService);
  
  I18nService.$inject = ['$cookies', '$window'];
  
  function I18nService($cookies, $window) {
    
    var cookieOptions = {'path': '/'};
       
    var messages = $.extend(true, {}, eval('locale_messages'));    
    var defaultLanguage = 'en-US';
    var supportLanguages = {
      'en-US': 'English',
      'zh-CN': '中文'
    };
    var isSupportLanguage = function(language) {
      for (var i in supportLanguages) {
        if(language === String(i)) {
          return true;
        }
      }
      return false;
    };
    
        
    return tr;
    function tr() {
      
      return {
        'setCurrentLanguage': function(language) {
          if(!angular.isDefined(language) || !isSupportLanguage(language)) {
            language = defaultLanguage;
          }
          $cookies.put('language', language, cookieOptions);
        },
        'setDefaultLanguage': function() {
          $cookies.put('language', defaultLanguage, cookieOptions);
        },
        'getCurrentLanguage': function() {
          return $cookies.get('language') || defaultLanguage;
        },
        'getLanguageName': function(language) {
          if(!angular.isDefined(language) || !isSupportLanguage(language)) {
            language = defaultLanguage;
          }
          $cookies.put('language', language, cookieOptions);
          return supportLanguages[language];    
        },
        'getSupportLanguages': function() {
          return supportLanguages;
        },
        'unset': function(){
          $cookies.put('language', defaultLanguage, cookieOptions);
        },
        'getValue': function(key) {
          return messages[key];
        }
      };
      
    }
  }
  
})();