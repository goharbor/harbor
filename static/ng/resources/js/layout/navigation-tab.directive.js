(function() {

  'use strict';

  angular
    .module('harbor.app')
    .directive('navigationTab', navigationTab);
  
  NavigationTabController.$inject = ['$window'];
    
  function NavigationTabController($window) {
    var vm = this;
    vm.location = $window.location.pathname;
    vm.closePane = closePane;
    function closePane() {
      vm.visible = false;
    }
  }
  
  function navigationTab() {
    var directive = {
      restrict: 'E',
      templateUrl: getTemplateUrl,
      link: link,
      scope: {
        templateUrl: "@",
        visible: "="
      },
      replace: true,
      controller: NavigationTabController,
      controllerAs: 'vm',
      bindToController: true
    }
    
    return directive;
    
    function getTemplateUrl(element, attrs) {
      return '/static/ng/resources/js/layout/'+ attrs.templateUrl;
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
        $(event.target).not('span').addClass('active');
      }
     
    }
   
  }
  
})();