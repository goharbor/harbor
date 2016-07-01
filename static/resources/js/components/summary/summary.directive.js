(function() {
    
  'use strict';
  
  angular
    .module('harbor.summary')
    .directive('projectSummary', projectSummary);
    
  ProjectSummaryController.$inject = ['$scope', 'StatProjectService', '$filter', 'trFilter'];
  
  function ProjectSummaryController($scope, StatProjectService, $filter, trFilter) {
    var vm = this;
    
    StatProjectService()
      .success(statProjectSuccess)
      .error(statProjectFailed);
      
    function statProjectSuccess(data) {
      vm.statProjects = data;
    }
    
    function statProjectFailed(data) {
      $scope.$emit('modalTitle', $filter('tr')('error'));
      $scope.$emit('modalMessage', $filter('tr')('failed_get_stat') + data);
      $scope.$emit('raiseError', true);
      console.log('Failed get stat:' + data);
    }
  }
  
  function projectSummary() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/summary/summary.directive.html',
      'controller': ProjectSummaryController,
      'scope' : true,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
  }
    
})();