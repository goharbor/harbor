(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .directive('retrieveProjects', retrieveProjects);
 
  function RetrieveProjectsController() {
    var vm = this;
   
    vm.selectItem = selectItem;
    vm.filterProjects = filterProjects;
    
    function selectItem(item) {
       vm.selectedId = item.id;
       vm.selectedProject = item;
    }
    
    var totalProjects = vm.projects;
    
    function filterProjects(input) {
     
      if(input == "") {
        vm.projects = totalProjects;
      }else{
        var filteredResults = [];
        for(var i = 0; i < totalProjects.length; i++) {
          var item = totalProjects[i];
          if(item.name.indexOf(input) >= 0) {
            filteredResults.push(item);
            continue;
          }
        }
        vm.projects = filteredResults;
      }
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
      replace: true,
      controller: RetrieveProjectsController,
      bindToController: true,
      controllerAs: 'vm'
    }
    
    return directive;
    
    
  }
  
})();