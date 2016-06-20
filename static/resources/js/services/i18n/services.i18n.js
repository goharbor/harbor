(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.i18n')
    .factory('I18nService', I18nService);
  
  I18nService.$inject = ['$cookies', '$window'];
  
  function I18nService($cookies, $window) {
    var cookieOptions = {'path': '/'};
    var messages = $.extend(true, {}, eval('locale_messages'));    
    var defaultLanguage = navigator.language || 'en-US';
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