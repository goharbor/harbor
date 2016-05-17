(function() {
  
  'use strict';
  
  angular
    .module('harbor.user')
    .directive('toggleAdmin', toggleAdmin);
   
  ToggleAdminController.$inject = ['$scope', 'ToggleAdminService'];
  
  function ToggleAdminController($scope, ToggleAdminService) {
    var vm = this;
    
    vm.isAdmin = (vm.hasAdminRole == 1) ? true : false;
    vm.toggle = toggle;
    
    function toggle() {
      ToggleAdminService(vm.userId)
        .success(toggleAdminSuccess)
        .error(toggleAdminFailed);        
    }    
    
    function toggleAdminSuccess(data, status) {
      if(vm.isAdmin) {
        vm.isAdmin = false;
      }else{
        vm.isAdmin = true;
      }
      console.log('Toggled userId:' + vm.userId + ' to admin:' + vm.isAdmin);
    }

    function toggleAdminFailed(data, status) {
      console.log('Failed toggle admin:' + data);
    }    
  }
  
  function toggleAdmin() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/user/toggle-admin.directive.html',
      'scope': {
        'hasAdminRole': '=',
        'userId': '@'
      },
      'link': link,
      'controller': ToggleAdminController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      
    }
  }
  
})();