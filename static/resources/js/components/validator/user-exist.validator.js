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