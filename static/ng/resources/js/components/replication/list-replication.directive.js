(function() {
  
  'use strict';
  
  angular
    .module('harbor.replication')
    .directive('listReplication', listReplication);
    
  ListReplicationController.$inject = ['ListReplicationPolicyService'];
  
  function ListReplicationController(ListReplicationPolicyService) {
    var vm = this;
    
    vm.addReplication = addReplication;
    vm.retrieve = retrieve;
    vm.last = false;
    vm.retrieve();
        
    function retrieve() {
      ListReplicationPolicyService()
        .then(listReplicationPolicySuccess, listReplicationPolicyFailed);
    }

    function listReplicationPolicySuccess(data, status) {
      vm.replicationPolicies = data || [];
    }
    
    function listReplicationPolicyFailed(data, status) {
      console.log('Failed list replication policy:' + data);
    }

    function addReplication() {
      vm.modalTitle = 'Create New Policy';
      vm.modalMessage = '';
    }
    
  }
  
  function listReplication($timeout) {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/ng/resources/js/components/replication/list-replication.directive.html',
      'scope': true,
      'link': link,
      'controller': ListReplicationController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      var uponPaneHeight = element.find('#upon-pane').height();
      var downPaneHeight = element.find('#down-pane').height() + element.find('#down-pane').offset().top;
      var handleHeight = element.find('.split-handle').height() + element.find('.split-handle').offset().top;
      
      console.log('uponPaneHeight:' + uponPaneHeight + ', downPaneHeight:' + downPaneHeight + ', handleHeight:' + handleHeight);
      
      element.find('.split-handle').on('mousedown', mousedownHandler);
      
      function mousedownHandler(e) {
        e.preventDefault();
        console.log('pageY:' + e.pageY + ', offset:' + (handleHeight - e.pageY));
        $(document).on('mousemove', mousemoveHandler);    
        $(document).on('mouseup', mouseupHandler);
      }
      
      function mousemoveHandler(e) {
        element.find('#upon-pane').css({'height' : uponPaneHeight - (handleHeight - e.pageY) + 'px'});
        element.find('#down-pane').css({'height' : downPaneHeight - (e.pageY - handleHeight)  + 'px'});
      }
      
      function mouseupHandler(e) {
        $(document).off('mousedown');
        $(document).off('mousemove');
      }
      
      scope.$watch('vm.last', function(current) { 
        $timeout(function(){
          if(current) {
            element.find('#upon-pane table>tbody>tr').on('click', trClickHandler);
            element.find('#upon-pane table>tbody>tr:eq(0)').trigger('click');
          }
        });
      });
      
      function trClickHandler(e) {
        element
          .find('#upon-pane table>tbody>tr')  
          .css({'background-color': '#FFFFFF'})
          .css({'color': '#000'});
        $(this)
          .css({'background-color': '#057ac9'})
          .css({'color': '#fff'});
      }
    }
  }
  
})();