(function() {
  
  'use strict';
  
  angular
    .module('harbor.repository')
    .directive('pullCommand', pullCommand);
  
  function PullCommandController() {
    
  }
  
  function pullCommand() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/repository/pull-command.directive.html',
      'scope': {
        'repoName': '@',
        'tag': '@'
      },
      'link': link,
      'controller': PullCommandController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
       
      ctrl.harborRegUrl = $('#HarborRegUrl').val() + '/';
    
      element.find('a').on('click', clickHandler);
      function clickHandler(e) {
        element.find('input[type="text"]').select();
      }
  
    }
    
  }
  
})();