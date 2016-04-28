(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .directive('retrieveProjects', retrieveProjects);
  
  RetrieveProjectsController.$inject = ['$scope', 'nameFilter'];
   
  function RetrieveProjectsController($scope, nameFilter) {
    var vm = this;
   
    $scope.$watch('vm.selectedProject', function(current, origin) {
      if(current) {        
        vm.selectedId = current.ProjectId;    
      }
    });
    
    vm.filterInput = "";
    vm.selectItem = selectItem;
    
    function selectItem(item) {
      vm.selectedId = item.ProjectId;
      vm.selectedProject = item;
      vm.isOpen = false;
    }       
    
  }
  
  function retrieveProjects() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/ng/resources/js/components/details/retrieve-projects.directive.html',
      scope: {
        'projects': '=',
        'isOpen': '=',
        'selectedProject': '='
      },
      link: link,
      replace: true,
      controller: RetrieveProjectsController,
      bindToController: true,
      controllerAs: 'vm'
    }
    
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      
    }
  }
  
})();