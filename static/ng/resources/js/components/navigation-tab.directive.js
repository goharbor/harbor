(function() {

  'use strict';

  angular
    .module('harbor.app')
    .directive('navigationTab', navigationTab);
    
  function navigationTab() {
    var directive = {
      restrict: 'EA',
      scope: {
        "tabs": "="
      },
    
      templateUrl: '/static/ng/resources/js/components/navigation-tab.directive.html',
      link: link
    }
    
    return directive;
    
    function link(scope, element, attrs) {
      element
        .on('mouseover', mouseover)
        .on('mouseout', mouseout);
      
      function mouseover(event) {
        $(event.target).addClass('active');
      }
      
      function mouseout(event) {
        $(event.target).removeClass('active');
      }
    }
  }
  
})();