(function() {

  'use strict';

  angular
    .module('harbor.validator')
    .directive('normalName', normalName); //仅限英文的标记名

  normalName.$inject = ['TAG_REGEXP'];

  function normalName(TAG_REGEXP) {
    var directive = {
      'require': 'ngModel',
      'link': link
    };
    return directive;

    function link(scope, element, attrs, ctrl) {
      ctrl.$validators.tagName = validator;

      function validator(modelValue, viewValue) {
        return TAG_REGEXP.test(modelValue);
      }
    }
  }

})();
