(function() {

  'use strict';

  angular
    .module('harbor.app')
    .directive('navigationTab', navigationTab);
    
  function navigationTab() {
    var directive = {
      restrict: 'E',
      templateUrl: getTemplateUrl,
      link: link,
      controller: controller
    }
    
    return directive;
    
    function getTemplateUrl(element, attrs) {
      return '/static/ng/resources/js/components/'+ attrs.templateUrl;
    }
    
    function link(scope, element, attrs, ctrl) {
     
      if (attrs.templateUrl.indexOf("navigation-tab") >= 0) {
        element.find('a[href$="' + ctrl.location + '"]').addClass('active');
      }
    
      if (attrs.templateUrl.indexOf("navigation-details") >= 0) {
        element.find('a:first').addClass('active');
      }
      
      element.on('click', click);
      
      function click(event) {
        element.find('a').removeClass('active');
        $(event.target).addClass('active');
      }
     
    }
    
    controller.$inject = ['$window'];
    
    function controller($window) {
      var vm = this;
      vm.location = $window.location.pathname;
    }
  }
  
})();