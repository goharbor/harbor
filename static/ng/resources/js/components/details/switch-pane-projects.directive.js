(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .directive('switchPaneProjects', switchPaneProjects);

  SwitchPaneProjectsController.$inject = ['$scope'];

  function SwitchPaneProjectsController($scope) {
    var vm = this;
 
    $scope.$watch('vm.selectedProject', function(current, origin) {
      if(current){
        vm.projectName = current.name;
        vm.selectedProject = current;
      }
    });
      
    vm.switchPane = switchPane;
    
    function switchPane() {
      if(vm.isOpen) {
        vm.isOpen = false;
      }else{
        vm.isOpen = true;
      }
    }
    
  }
  
  function switchPaneProjects() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/ng/resources/js/components/details/switch-pane-projects.directive.html',
      scope: {
        'isOpen': '=',
        'selectedProject': '='
      },
      controller: SwitchPaneProjectsController,
      controllerAs: 'vm',
      bindToController: true
    };
    
    return directive;
      
  }
  
})();