(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .directive('switchPaneProjects', switchPaneProjects);

  function SwitchPaneProjectsController() {
    var vm = this;
    vm.projectName = "myrepo1";
    vm.open = false;    
    vm.switchPane = switchPane;
    
    function switchPane() {
      if(vm.open) {
        vm.open = false;
      }else{
        vm.open = true;
      }
 console.log(vm.open);
    }
  }
  
  function switchPaneProjects() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/ng/resources/js/components/details/switch-pane-projects.directive.html',
      replace: true,
      scope: {
        'projectName': '@',
        'open': '='
      },
      controller: SwitchPaneProjectsController,
      controllerAs: 'vm',
      bindToController: true
    }
    
    return directive;
  
  }
  
})();