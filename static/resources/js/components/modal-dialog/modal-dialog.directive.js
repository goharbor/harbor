(function() {
  
  'use strict';
  
  angular
    .module('harbor.modal.dialog')
    .directive('modalDialog', modalDialog);
  
  ModalDialogController.$inject = ['$scope'];
  
  function ModalDialogController($scope) {
    var vm = this;
  }
  
  function modalDialog() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/modal-dialog/modal-dialog.directive.html',
      'link': link,
      'scope': {
        'contentType': '@',
        'title': '@',
        'message': '@',
        'action': '&'
      },
      'controller': ModalDialogController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      
      if(!angular.isDefined(ctrl.contentType)) {
        ctrl.contentType = 'text/plain';  
      }
      console.log('Received contentType in modal:' + ctrl.contentType);
                  
      scope.$watch('vm.message', function(current) {
        if(current) {
          switch(ctrl.contentType) {
          case 'text/html':
            element.find('.modal-body').html(current); break;
          case 'text/plain':
            element.find('.modal-body').text(current); break;
          default:
            element.find('.modal-body').text(current); break;
          }
        }
      });
      
      element.find('#btnOk').on('click', clickHandler);

      function clickHandler(e) {
        ctrl.action();
        element.find('#myModal').modal('hide');
      }
    }
  }
  
})();