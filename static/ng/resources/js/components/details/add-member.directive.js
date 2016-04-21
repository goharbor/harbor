(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .directive('addMember', addMember);
    
  function AddMemberController() {
    var vm = this;
    
  }
  
  function addMember() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/details/add-member.directive.html',
      'scope': {
        
      },
      'controller': AddMemberController,
      'controllerAs': 'vm',
      'bindToController': true
    }
    return directive;
  }
  
})();