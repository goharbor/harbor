(function() {
  
  'use strict';
  
  angular
    .module('harbor.layout.element.height')
    .directive('elementHeight', elementHeight);
    
  function elementHeight($window) {
    var directive = {
      'restrict': 'A',
      'link': link
    };
    
    return directive;
    
    function link(scope, element, attrs) {

      var w = angular.element($window);

      scope.getDimension = function() {
        return {'h' : w.height()};
      };
      
      if(!angular.isDefined(scope.subsHeight))  scope.subsHeight = 110;
      if(!angular.isDefined(scope.subsSection)) scope.subsSection = 32;
      if(!angular.isDefined(scope.subsSubPane)) scope.subsSubPane = 226;
      if(!angular.isDefined(scope.subsTblBody)) scope.subsTblBody = 40;
    
      scope.$watch(scope.getDimension, function(current) {
        if(current) {
          var h = current.h; 
          element.find('.section').css({'height': (h - scope.subsHeight - scope.subsSection) + 'px'});        
          element.find('.sub-pane').css({'height': (h - scope.subsHeight - scope.subsSubPane) + 'px'});
          element.find('.tab-pane').css({'height': (h - scope.subsHeight - scope.subsSubPane) + 'px'});
          var subPaneHeight = element.find('.sub-pane').height();
          element.find('.table-body-container').css({'height': (subPaneHeight - scope.subsTblBody) + 'px'});
        }
      }, true);
      
      w.on('pageshow, resize', function() {       
        scope.$apply();
      });
    }
  }
  
})();