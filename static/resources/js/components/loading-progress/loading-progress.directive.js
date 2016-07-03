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
        'hideTarget': '@'
      },
      'link': link
    };
    
    return directive;
    
    function link(scope, element, attrs) {
      var spinner = $('<span class="loading-progress">');

      function convertToBoolean(val) {
        return val === 'true' ? true : false;
      }
      
      var hideTarget = convertToBoolean(scope.hideTarget);
      
      console.log('loading-progress, toggleInProgress:' + scope.toggleInProgress + ', hideTarget:' + hideTarget);
      
      var pristine = element.html();
                 
      scope.$watch('toggleInProgress', function(current) {
        if(scope.toggleInProgress) {
          element.attr('disabled', 'disabled');
          if(hideTarget) {
            element.html(spinner);
          }else{
            spinner = spinner.css({'margin-left': '5px'});
            element.append(spinner);
          }
        }else{
          if(hideTarget) {
            element.html(pristine);
          }else{
            element.find('.loading-progress').remove();
          }
          element.removeAttr('disabled');
        }
      });

    }
  }
  
})();