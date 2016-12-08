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
    .directive('charsLength', charsLength);
  
  charsLength.$inject = ['ASCII_CHARS'];
  
  function charsLength(ASCII_CHARS) {
    var directive = {
      'require': 'ngModel',
      'scope': {
        min: '@',
        max: '@'
      },
      'link': link
    };
    
    return directive;
    
    function link(scope, element, attrs, ctrl) {
        
      ctrl.$validators.charsLength = validator;
      
      function validator(modelValue, viewValue) {
        if(ctrl.$isEmpty(modelValue)) {
          return true;
        }
        
        var actualLength = 0;
        
        if(ASCII_CHARS.test(modelValue)) {
          actualLength = modelValue.length;
        }else{
          for(var i = 0; i < modelValue.length; i++) {
            ASCII_CHARS.test(modelValue[i]) ? actualLength += 1 : actualLength += 2;
          }
        }
                
        if(attrs.min && actualLength < attrs.min) {
          return false;
        }
        
        if(attrs.max && actualLength > attrs.max) {
          return false;
        }
        
        return true;
      }
      
    }
  }
  
  
})();