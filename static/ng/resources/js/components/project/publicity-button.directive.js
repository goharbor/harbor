(function() {
  
  'use strict';
  
  angular
    .module('harbor.project')
    .directive('publicityButton', publicityButton);
  
  PublicityButtonController.$inject = ['ToggleProjectPublicityService'];
  
  function PublicityButtonController(ToggleProjectPublicityService) {
    var vm = this;
    vm.toggle = toggle;
    
    if(vm.isPublic === 1) {
      vm.isPublic = true;
    }else{
      vm.isPublic = false;
    }
        
    function toggle() {
      
      if(vm.isPublic) {
        vm.isPublic = false;
      }else{
        vm.isPublic = true;
      }
      
      ToggleProjectPublicityService(vm.projectId, vm.isPublic)
        .success(toggleProjectPublicitySuccess)
        .error(toggleProjectPublicityFailed);
    }
    
    function toggleProjectPublicitySuccess(data, status) {
      console.log('Successful toggle project publicity.');
    }
    
    function toggleProjectPublicityFailed(e) {
      console.log('Failed toggle project publicity:' + e);
    }
  }

  function publicityButton() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/project/publicity-button.directive.html',
      'scope': {
        'isPublic': '=',
        'owned': '=',
        'projectId': '='
      },
      'link': link,
      'controller': PublicityButtonController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attr, ctrl) {
      
    }
  }
  
})();