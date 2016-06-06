(function() {
    
    'use strict';
    
    angular
      .module('harbor.summary')
      .directive('projectSummary', projectSummary);
      
    ProjectSummaryController.$inject = ['StatProjectService', 'getStatisticsName'];
    
    function ProjectSummaryController(StatProjectService, getStatisticsName) {
        var vm = this;
        vm.getSummaryName = getSummaryName;
        
        StatProjectService()
          .success(statProjectSuccess)
          .error(statProjectFailed);
          
        function statProjectSuccess(data, status) {
            vm.statProjects = data;
        }
        
        function statProjectFailed(status) {
            console.log('Failed stat project:' + status);
        }
        
        function getSummaryName(payloadName) {
           var statisticsName =  getStatisticsName({'key': 'payloadName', 'value': payloadName}); 
           return statisticsName.name;
       }           
    }
    
    function projectSummary() {
        var directive = {
          'restrict': 'E',
          'templateUrl': '/static/ng/resources/js/components/summary/summary.directive.html',
          'controller': ProjectSummaryController,
          'scope' : true,
          'controllerAs': 'vm',
          'bindToController': true
        };
        
        return directive;
    }
    
})();