(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.navigation')
    .directive('navigationAdminOptions', navigationAdminOptions);
  
  NavigationAdminOptions.$inject = ['$location'];
  
  function NavigationAdminOptions($location) {
    var vm = this;
    vm.path = $location.path();
  }
  
  function navigationAdminOptions() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/layout/navigation/navigation-admin-options.directive.html',
      'scope': {
        'target': '='
      },
      'link': link,
      'controller': NavigationAdminOptions,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      var visited = ctrl.path.substring(1);  
      console.log('visited:' + visited);

      if(visited) {
        element.find('a[tag="' + visited + '"]').addClass('active');
      }else{
        element.find('a:first').addClass('active');
      }
      
      element.find('a').on('click', click);
            
      function click(event) {
        element.find('a').removeClass('active');
        $(event.target).addClass('active');
        ctrl.target = $(this).attr('tag');
        scope.$apply();
      }
    }
  }
  
})();