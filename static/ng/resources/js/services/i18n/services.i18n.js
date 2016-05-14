(function() {
  
  'use strict';
  
  angular
    .module('harbor.services.i18n')
    .factory('I18nService', I18nService);
  
  I18nService.$inject = ['$cookies', '$window'];
  
  function I18nService($cookies, $window) {
    var languages = $.extend(true, {}, global_messages);
    var defaultLanguage = navigator.language || 'en-US';
    var languageNames = {
      'en-US': 'English',
      'zh-CN': '中文'
    };    
    return tr;
    function tr() {
      return {
        'setCurrentLanguage': function(language) {          
          if(!language){
            language = defaultLanguage;
          }
          $cookies.put('language', language, {'path': '/ng/'});
          $window.location.reload();
        },
        'getCurrentLanguage': function() {
          return $cookies.get('language') || defaultLanguage;
        },
        'getLanguageName': function(crrentLanguage) {
          return languageNames[crrentLanguage];    
        },
        'getValue': function(key, currentLanguage) {
          return languages[key][currentLanguage];
        }
      }
    }
  }
  
})();