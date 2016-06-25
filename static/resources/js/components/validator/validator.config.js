(function() {
  
  'use strict';
  
  angular
    .module('harbor.validator')
    .constant('INVALID_CHARS', [",","~","#", "$", "%"])
    .constant('PASSWORD_REGEXP', /^(?=.*\d)(?=.*[a-z])(?=.*[A-Z])(?!.*\s).{7,20}$/)
    .constant('PROJECT_REGEXP', /^[a-z0-9](?:-*[a-z0-9])*(?:[._][a-z0-9](?:-*[a-z0-9])*)*$/);
})();