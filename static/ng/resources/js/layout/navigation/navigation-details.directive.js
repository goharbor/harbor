(function() {

  'use strict';

  angular
    .module('harbor.layout.navigation')
    .directive('navigationDetails', navigationDetails);
  
  NavigationDetailsController.$inject = ['$window', '$location', '$scope'];
  
  function NavigationDetailsController($window, $location, $scope) {
    var vm = this;    
    
    $scope.$watch('vm.selectedProject', function(current, origin) {
      if(current) {
        vm.projectId = current.ProjectId;
      }
    });
    
    vm.url = $location.url();
    vm.clickTab = clickTab;
     
    function clickTab() {       
      console.log("triggered clickTab of Controller.");
      vm.isOpen = false;  
      $scope.$apply();
    }
 
  }
  
  function navigationDetails() {
    var directive = {
      restrict: 'E',
      templateUrl: '/static/ng/resources/js/layout/navigation/navigation-details.directive.html',
      link: link,
      scope: {
        'isOpen': '=',
        'selectedProject': '='
      },
      replace: true,
      controller: NavigationDetailsController,
      controllerAs: 'vm',
      bindToController: true
    }
    
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      
      var visited = ctrl.url.substring(1, ctrl.url.indexOf('?') - 1);
     
      element.find('a[tag^="' + visited + '"]').addClass('active');
      element.on('click', click);
      
      function click(event) {
        element.find('a').removeClass('active');
        $(event.target).not('span').addClass('active');
        ctrl.clickTab();
      }
     
    }
   
  }
  
})();