(function() {
  
  'use strict';
  
  angular
    .module('harbor.user')
    .directive('toggleAdmin', toggleAdmin);
   
  ToggleAdminController.$inject = ['$scope', 'ToggleAdminService', '$filter', 'trFilter'];
  
  function ToggleAdminController($scope, ToggleAdminService, $filter, trFilter) {
    var vm = this;
    
    vm.isAdmin = (vm.hasAdminRole == 1) ? true : false;
    vm.enabled = vm.isAdmin ? 0 : 1;
    vm.toggle = toggle;
    
    function toggle() {
      ToggleAdminService(vm.userId, vm.enabled)
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
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_toggle_admin'));
      $scope.$emit('raiseError', true);
      console.log('Failed to toggle admin:' + data);
    }    
  }
  
  function toggleAdmin() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/user/toggle-admin.directive.html',
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