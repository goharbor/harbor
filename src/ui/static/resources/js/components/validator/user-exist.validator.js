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
    .directive('userExists', userExists);
  
  userExists.$inject = ['UserExistService'];
  
  function userExists(UserExistService) {
    var directive = {
      'require': 'ngModel',
      'scope': {
        'target': '@'
      },
      'link': link
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      
      var valid = true;     
                  
      ctrl.$validators.userExists = validator;
      
      function validator(modelValue, viewValue) {
      
        console.log('modelValue:' + modelValue + ', viewValue:' + viewValue);
         
        if(ctrl.$isEmpty(modelValue)) {
          console.log('Model value is empty.');
          return true;
        }

        UserExistService(attrs.target, modelValue)
          .success(userExistSuccess)
          .error(userExistFailed);   
            
        function userExistSuccess(data, status) {
          valid = !data;
          if(!valid) {
            console.log('Model value already exists');
          }
        }
        
        function userExistFailed(data, status) {
          console.log('Failed to in retrieval:' + data);
        }       
        
        return valid;
      }  
    }
    
  }
})();