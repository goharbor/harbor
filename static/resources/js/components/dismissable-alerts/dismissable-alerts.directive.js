(function() {
  
  'use strict';
  
  angular
    .module('harbor.dismissable.alerts')
    .directive('dismissableAlerts', dismissableAlerts);

  function dismissableAlerts() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/dismissable-alerts/dismissable-alerts.directive.html',
      'link': link
    };
    return directive;
    function link(scope, element, attrs, ctrl) {
      
      scope.close = function() {
        scope.toggleAlert = false;
      }
      scope.$on('raiseAlert', function(e, val) {
        console.log('received raiseAlert:' + angular.toJson(val));
        if(val.show) {
          scope.message = val.message;
          scope.toggleAlert = true;
        }else{
          scope.message = ''
          scope.toggleAlert = false;
        }
      });
    }
  }
  
})();