(function() {
  
  'use strict';
  
  angular
    .module('harbor.project')
    .directive('publicityButton', publicityButton);
  
  PublicityButtonController.$inject = ['EditProjectService'];
  
  function PublicityButtonController(EditProjectService) {
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
      
      EditProjectService(vm.projectId, vm.isPublic)
        .success(editProjectSuccess)
        .error(editProjectFailed);
    }
    
    function editProjectSuccess(data, status) {
      console.log('edit project successfully:' + status);
    }
    
    function editProjectFailed(e) {
      console.log('edit project failed:' + e);
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