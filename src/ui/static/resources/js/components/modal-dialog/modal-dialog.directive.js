/*
    Copyright (c) 2016 VMware, Inc. All Rights Reserved.
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
        
        http://www.apache.org/licenses/LICENSE-2.0
        
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/
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
        'modalTitle': '@',
        'modalMessage': '@',
        'action': '&',
        'confirmOnly': '='
      },
      'controller': ModalDialogController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
           
      scope.$watch('contentType', function(current) {
        if(current) {
          ctrl.contentType = current;  
        }
      });
      scope.$watch('confirmOnly', function(current) {
        if(current) {
          ctrl.confirmOnly = current;
        }
      });
                        
      scope.$watch('vm.modalMessage', function(current) {
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
      
      scope.$on('showDialog', function(e, val) {
        if(val) {
          element.find('#myModal').modal('show');
        }else{
          element.find('#myModal').modal('hide');
        }
      });
        
      element.find('#btnOk').on('click', clickHandler);        

      function clickHandler(e) {
        ctrl.action();  
      }
    }
  }
  
})();
