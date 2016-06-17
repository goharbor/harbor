(function() {
  
  'use strict';
  
  angular
    .module('harbor.validator')
    .directive('compareTo', compareTo);
  
  function compareTo() {
    var directive = {
      'require' : 'ngModel',
      'scope':{
        'otherModelValue': '=compareTo'
      },
      'link': link
    };
    return directive;
    
    function link (scope, element, attrs, ctrl) {
      
      ctrl.$validators.compareTo = validator;
      
      function validator(modelValue) {
        return modelValue === scope.otherModelValue;
      }
      
      scope.$watch("otherModelValue", function(current, origin) {
        ctrl.$validate();
      });
    }    
  }
  
})();