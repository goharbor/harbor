(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.sign.up')
    .factory('validationConfig', validationConfig)
    .factory('Validator', Validator)
    .controller('SignUpController', SignUpController);
  
  function validationConfig() {
    return config;
    function config() {
      var invalidChars = [",","~","#", "$", "%"];
      var vc = {
        'username': {
          'required' : {'value': true, 'message': 'Username is required.'}, 
          'maxLength': {'value': 20, 'message': 'Maximum 20 characters.'},
          'invalidChars': {'value': invalidChars, 'message': 'Username contains invalid characters.'},
          'exists': {'value': false, 'message': 'Username already exists.'}
        },
        'email': {
          'required' : {'value': true, 'message': 'Email is required.'},
          'regexp': {'value': /^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/, 
                     'message': 'Email is invalid.'},
          'exists': {'value': false, 'message': 'Email address already exists.'}
        },
        'realname': {
          'maxLength': {'value': 20, 'message': 'Maximum 20 characters.'}
        },
        'password': {
          'required': {'value': true, 'message': 'Password is required.'},          
          'complexity': {'value':  /^(?=.*\d)(?=.*[a-z])(?=.*[A-Z])(?!.*\s).{7,20}$/, 
                         'message': 'At least 7 characters with 1 lowercase letter, 1 capital letter and 1 numeric character.'}
        },
        'confirmPassword': {
          'compareWith': {'value': true, 'message': 'Confirm password mismatch.'}
        }
      };
      return vc;      
    }
    
  }
  
  Validator.$inject = ['validationConfig', 'UserExistService'];
  
  var status = function(isValid, message) {
    this.isValid = isValid;
    this.message = message;
  }
  
  function Validator(validationConfig, UserExistService) {
    
    return validate;
    
    var valid = true;
    
    function validate(fieldName, fieldValue, options) {
      
      var config = validationConfig()[fieldName];
      
      console.log('Checking ' + fieldName + ' for value:' + fieldValue);
      
      for(var c in config) {
        console.log('item:' + c + ', criterion: ' + config[c]['value']);
        switch(c) {
        case 'required':
          valid = required(fieldValue); break;
        case 'maxLength':
          valid = maxLength(fieldValue, config[c]['value']); break;
        case 'regexp':
          valid = regexp(fieldValue, config[c]['value']); break;
        case 'invalidChars':
          valid = invalidChars(fieldValue, config[c]['value']); 
          break;
        case 'exists':
          exists(fieldName, fieldValue);
          break;
        case 'complexity':
          valid = complexity(fieldValue, config[c]['value']); break;
        case 'compareWith':
          valid = compareWith(fieldValue, options); break;
        }
        if(!valid) {
          return new status(valid, config[c]['message']);
        }
      }
      return new status(valid, '');
    }
    
    function required(value) {
      return (typeof value != "undefined" && value != null && $.trim(value).length > 0);
    }
    
    function maxLength(value, maximum) {
      console.log('maximum is:' + maximum); 
      return (value.length <= maximum);
    }
    
    function regexp(value, reg) {
      return reg.test(value);
    }
    
    function invalidChars(value, chars) {
      for(var i = 0 ; i < chars.length; i++) {
        var char = chars[i];
        if (value.indexOf(char) >= 0) {
          return false;
        }
      }
      return true;
    }
    
    function exists(target, value) {
      UserExistService(target, value)
        .success(userExistSuccess)
        .error(userExistFailed);          
    }
    
    function userExistSuccess(data, status) {
      valid = !data;
    }
    
    function userExistFailed(data, status) {
      console.log('error in checking user exists:' + data);
    }    
    
    function complexity(value, reg) {
      return reg.test(value);
    }
    
    function compareWith(value, comparedValue) {
      return value === comparedValue;
    }
   
  }
  
  SignUpController.$inject = ['SignUpService', 'Validator'];
 
  function SignUpController(SignUpService, Validator) {
    var vm = this;
    vm.username = 'user122';
    var status = Validator('username', vm.username);
    console.log('status, isValid:' + status.isValid + ', message:' + status.message);
  }
  
})();