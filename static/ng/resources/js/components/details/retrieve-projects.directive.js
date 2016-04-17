(function() {
  
  'use strict';
  
  angular
    .module('harbor.details')
    .constant('mockupProjects', mockupProjects)
    .directive('retrieveProjects', retrieveProjects);
  
  function mockupProjects() {
    var data = [
      { "id": 1, "name" : "myrepo"},
      { "id": 2, "name" : "myproject"},
      { "id": 3, "name" : "harbor_project"},
      { "id": 4, "name" : "legacy"} 
    ];
    return data;
  }
  
  RetrieveProjectsController.$inject = ['mockupProjects'];
 
  function RetrieveProjectsController(mockupProjects) {
    var vm = this;
    vm.projects = mockupProjects();
    vm.selectItem = selectItem;
    vm.filterProjects = filterProjects;
    
    function selectItem(item) {
       vm.selectedId = item.id;
    }
    
    var totalProjects = mockupProjects();
    
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
        'visible': '='
      },
      replace: true,
      controller: RetrieveProjectsController,
      bindToController: true,
      controllerAs: 'vm'
    }
    
    return directive;
    
    
  }
  
})();