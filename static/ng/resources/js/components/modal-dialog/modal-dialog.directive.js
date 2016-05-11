(function() {
  
  'use strict';
  
  angular
    .module('harbor.modal.dialog')
    .directive('modalDialog', modalDialog);
  
  function ModalDialogController() {
    var vm = this;
    vm.action();
  }
  
  function modalDialog() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/modal-dialog/modal-dialog.directive.html',
      'link': link,
      'scope': {
        'message': '@',
        'action': '&'
      },
      'controller': ModalDialogController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      
    }
  }
  
})();