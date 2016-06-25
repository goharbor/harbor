(function() {
  
  'use strict';
  
  angular
    .module('harbor.validator')
    .directive('projectName', projectName);
  
  projectName.$inject = ['PROJECT_REGEXP']
  
  function projectName(PROJECT_REGEXP) {
    var directive = {
      'require': 'ngModel',
      'link': link
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      ctrl.$validators.projectName = validator;
      
      function validator(modelValue, viewValue) {
        return PROJECT_REGEXP.test(modelValue);
      }
    }
  }
  
})();