(function() {
  
  'use strict';
  
  angular
    .module('harbor.inline.help')
    .directive('inlineHelp', inlineHelp);
  function InlineHelpController() {
    var vm = this;
  }
  function inlineHelp() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/inline-help/inline-help.directive.html',
      'scope': {
        'helpTitle': '@',
        'content': '@'
      },
      'link': link,
      'controller': InlineHelpController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    function link(scope, element, attr, ctrl) {
      element.popover({
        'title': ctrl.helpTitle,
        'content': ctrl.content,
        'html': true
      });
    }
  }
  
})();