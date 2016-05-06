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
     
      ZeroClipboard.config( { swfPath: "/static/ng/vendors/zc/v2.2.0/ZeroClipboard.swf" } );
      var clip = new ZeroClipboard(element.find('a'));
      element.find('span').tooltip({'trigger': 'click'});
      
      clip.on("ready", function() {
        console.log("Flash movie loaded and ready.");
        this.on("aftercopy", function(event) {
          console.log("Copied text to clipboard: " + event.data["text/plain"]);
          element.find('span').tooltip('show');
          setTimeout(function(){
            element.find('span').tooltip('hide');
          }, 1000);
        });
      });
  
      clip.on("error", function(event) {
        console.log('error[name="' + event.name + '"]: ' + event.message);
        ZeroClipboard.destroy();
        element.find('span').tooltip('destroy');
      });
  
    }
    
  }
  
})();