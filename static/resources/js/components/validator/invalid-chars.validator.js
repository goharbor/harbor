(function() {
  
  'use strict';
  
  angular
    .module('harbor.validator')
    .directive('invalidChars', invalidChars);
  
  invalidChars.$inject = ['INVALID_CHARS'];
  
  function invalidChars(INVALID_CHARS) {
    var directive = {
      'require': 'ngModel',
      'link': link
    };
    
    return directive;
    
    function link(scope, element, attrs, ctrl) {
        
      ctrl.$validators.invalidChars = validator;
      
      function validator(modelValue, viewValue) {
        if(ctrl.$isEmpty(modelValue)) {
          return true;
        }
      
        for(var i = 0; i < INVALID_CHARS.length; i++) {
          if(modelValue.indexOf(INVALID_CHARS[i]) >= 0) {
            return false;
          }
        }  
        
        return true;
      }
      
    }
  }
  
  
})();