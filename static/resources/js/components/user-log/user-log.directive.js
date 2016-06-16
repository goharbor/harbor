(function() {
    
    'use strict';
    
    angular
      .module('harbor.user.log')
      .directive('userLog', userLog);
      
    UserLogController.$inject = ['ListIntegratedLogService'];
    
    function UserLogController(ListIntegratedLogService) {
        var vm = this;
        
        ListIntegratedLogService()
          .success(listIntegratedLogSuccess)
          .error(listIntegratedLogFailed);
          
        function listIntegratedLogSuccess(data) {
            vm.integratedLogs = data || []
        }
   
        function listIntegratedLogFailed(data, status) {
            console.log('Failed list integrated logs:' + status);
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