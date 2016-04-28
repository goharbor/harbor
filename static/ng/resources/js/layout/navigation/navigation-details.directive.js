(function() {

  'use strict';

  angular
    .module('harbor.layout.navigation')
    .directive('navigationDetails', navigationDetails);
  
  NavigationDetailsController.$inject = ['$location', '$scope'];
  
  function NavigationDetailsController($location, $scope) {
    var vm = this;    
    vm.clickTab = clickTab;    
    vm.url = $location.url();
    
    if(vm.url == "/") {
      $location.url('/repositories');
    }
     
    function clickTab() { 
      vm.isOpen = false;  
      vm.url = $location.url();
      $scope.$emit('selectedProjectId', vm.selectedProject.ProjectId);
    }
 
  }
  
  function navigationDetails() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/ng/resources/js/layout/navigation/navigation-details.directive.html',
      link: link,
      scope: {
        'isOpen': '=',
        'selectedProject': "="
      },
      replace: true,
      controller: NavigationDetailsController,
      controllerAs: 'vm',
      bindToController: true
    }
    
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      
      var visited = ctrl.url;
     
      if(visited == "/") {
        element.find('a:first').addClass('active');
      }else{
        element.find('a[href$="' + visited + '"]').addClass('active');
      }
            
      element.on('click', click);
      
      
      
      
      function click(event) {
        element.find('a').removeClass('active');
        $(event.target).not('span').addClass('active');
        
        ctrl.clickTab();
      }
     
    }
   
  }
  
})();