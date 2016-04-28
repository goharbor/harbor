(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .directive('switchPaneProjects', switchPaneProjects);

  SwitchPaneProjectsController.$inject = ['$scope'];

  function SwitchPaneProjectsController($scope) {
    var vm = this;
    
    $scope.$on('isOpen', function(e, val){
      vm.isOpen = val;
    });
        
//    $scope.$watch('vm.selectedProject', function(current, origin) {
//      if(current){
//        vm.projectName = current.Name;   
//      }
//    });
      
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
      replace: true,
      scope: {
        'selectedProject': '=',
        'isOpen': '='
      },
      link: link,
      controller: SwitchPaneProjectsController,
      controllerAs: 'vm',
      bindToController: true
    }
    
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      scope.$watch('vm.selectedProject', function(current, origin) {
        if(current){
          scope.$emit('selectedProjectId', current.ProjectId);
          ctrl.projectName = current.Name;
        }
      });
    }
  
  }
  
})();