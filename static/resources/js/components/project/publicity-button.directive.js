(function() {
  
  'use strict';
  
  angular
    .module('harbor.project')
    .directive('publicityButton', publicityButton);
  
  PublicityButtonController.$inject = ['$scope', 'ToggleProjectPublicityService', '$filter', 'trFilter'];
  
  function PublicityButtonController($scope, ToggleProjectPublicityService, $filter, trFilter) {
    var vm = this;
    vm.toggle = toggle;
    
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
    
    function toggleProjectPublicityFailed(e, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      var message;
      if(status === 403) {
        message = $filter('tr')('failed_to_toggle_publicity_insuffient_permissions');
      }else{
        message = $filter('tr')('failed_to_toggle_publicity');
      }
      $scope.$emit('modalMessage', message);
      $scope.$emit('raiseError', true);
      
      if(vm.isPublic) {
        vm.isPublic = false;
      }else{
        vm.isPublic = true;
      }
      
      console.log('Failed to toggle project publicity:' + e);
    }
  }

  function publicityButton() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/project/publicity-button.directive.html',
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
      scope.$watch('vm.isPublic', function(current, origin) {
        if(current) {
          ctrl.isPublic = current;
        }
      });  
    }
  }
  
})();