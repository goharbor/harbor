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