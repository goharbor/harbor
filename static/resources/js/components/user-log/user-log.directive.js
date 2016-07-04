(function() {
    
  'use strict';
  
  angular
    .module('harbor.user.log')
    .directive('userLog', userLog);
    
  UserLogController.$inject = ['$scope', 'ListIntegratedLogService', '$filter', 'trFilter'];
  
  function UserLogController($scope, ListIntegratedLogService, $filter, trFilter) {
    var vm = this;
    
    ListIntegratedLogService()
      .success(listIntegratedLogSuccess)
      .error(listIntegratedLogFailed);
      
    function listIntegratedLogSuccess(data) {
      vm.integratedLogs = data || []
    }

    function listIntegratedLogFailed(data, status) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_to_get_user_log') + data);
      $scope.$emit('raiseError', true);
      console.log('Failed to get user logs:' + data);
    }
  }
  
  function userLog() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/user-log/user-log.directive.html',
      'controller': UserLogController,
      'scope' : true,
      'controllerAs': 'vm',
      'bindToController': true
    };
    
    return directive;
  }
    
})();