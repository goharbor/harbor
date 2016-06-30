(function() {
  
  'use strict';
  
  angular
    .module('harbor.loading.progress')
    .directive('loadingProgress', loadingProgress);
  
  function loadingProgress() {
    var directive = {
      'restrict': 'EA',
      'scope': {
        'toggleInProgress': '=',
        'hideTarget': '='
      },
      'link': link
    };
    
    return directive;
    
    function link(scope, element, attrs) {
      var spinner = $('<div>')
        .css({'display': 'inline-block'})
        .css({'position': 'relative'}) 
        .css({'background-image': 'url(/static/resources/img/loading.gif)'}) 
        .css({'background-position': 'center'})
        .css({'background-size': '107px'})
        .css({'width': '1.2em'})
        .css({'height': '1.2em'})
        .css({'margin': '0 0 1px 8px'})
        .css({'vertical-align': 'middle'});  
      
      scope.$watch('toggleInProgress', function(current) {
        console.log('toggleInProgress:' + scope.toggleInProgress);
        if(scope.toggleInProgress) {
          element.append(spinner);
          element.parent().attr('disabled', 'disabled');
          if(scope.hideTarget) {
            element.append(spinner);
            element.hide();     
          }
        }else{
          scope.hideTarget = false;
          element.show();
          element.find('div').remove();
          element.removeAttr('disabled');
        }
      });

    }
  }
  
})();