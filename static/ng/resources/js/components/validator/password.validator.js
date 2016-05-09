(function() {
  
  'use strict';
  
  angular
    .module('harbor.validator')
    .directive('password', password);
  
  password.$inject = ['PASSWORD_REGEXP'];
    
  function password(PASSWORD_REGEXP) {
    var directive = {
      'require' : 'ngModel',
      'link': link
    };
    return directive;
    
    function link (scope, element, attrs, ctrl) {
      
      ctrl.$validators.password = validator;
           
      function validator(modelValue, viewValue) {
        
        return PASSWORD_REGEXP.test(modelValue);
          
      }
    }
    
    
    
  }
  
})();