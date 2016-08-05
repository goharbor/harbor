(function() {
  
  'use strict';
  
  angular
    .module('harbor.paginator')
    .directive('paginator', paginator);
    
  PaginatorController.$inject = [];
  
  function PaginatorController() {
    var vm = this;   
  }
  
  paginator.$inject = [];
  
  function paginator() {
    var directive = {
      'restrict': 'E',
      'templateUrl': '/static/resources/js/components/paginator/paginator.directive.html',
      'scope': {
        'totalCount': '@',
        'pageSize': '@',
        'page': '=',
        'displayCount': '@',
        'retrieve': '&'
      },
      'link': link,
      'controller': PaginatorController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
     
      scope.$watch('vm.totalCount', function(current) {
        if(current) {
          var totalCount   = current;         
          var pageSize     = parseInt(ctrl.pageSize);
          var displayCount = parseInt(ctrl.displayCount);
          
          console.log('Total Count:' + totalCount + ', Page Size:' + pageSize + ', Display Count:' + displayCount);
          
          var TimeCounter = function() {
            this.time = 0;
            this.minimum = 0;
            this.maximum = 0;
          }
          
          TimeCounter.prototype.setMaximum = function(maximum) {
            this.maximum = maximum;
          }
          
          TimeCounter.prototype.increment = function() {
            if(this.time < this.maximum) {
              ++this.time;
              ++ctrl.page;
            }
          }
          
          TimeCounter.prototype.canIncrement = function() {
            if(this.time < this.maximum) {
              return true;
            }
            return false;
          }
          
          TimeCounter.prototype.decrement = function() {
            if(this.time > this.minimum) {
              --this.time;
              --ctrl.page;
            }
          }
          
          TimeCounter.prototype.canDecrement = function() {
            if(this.time > this.minimum) {
              return true;
            }
            return false;
          }
          
          TimeCounter.prototype.getTime = function() {
            return this.time;
          }
          
          var buttonCount = Math.ceil(totalCount / pageSize);
          var tc = new TimeCounter();
                      
          if(buttonCount <= displayCount) {
            tc.setMaximum(0);
          }else{
            tc.setMaximum(Math.floor(buttonCount / displayCount));
          }
         
          element.find('ul li:first a').on('click', previous);
          ctrl.showPrevious = false;
          
          element.find('ul li:last a').on('click', next);
          ctrl.showNext = (buttonCount > displayCount);
          
          var drawButtons = function(time) {
            element.find('li[tag="pagination-button"]').remove();
            var buttons = [];
            for(var i = 1; i <= displayCount; i++) {
              var displayNumber = displayCount * time + i;
              if(displayNumber <= buttonCount) {
                buttons.push('<li tag="pagination-button"><a href="javascript:void(0)" page="' + displayNumber + '">' + displayNumber + '</a></li>');
              }
            }
            $(buttons.join(''))
              .insertAfter(element.find('ul li:eq(0)')).end()
              .on('click', buttonClickHandler); 
          }
          
          drawButtons(tc.getTime());    
          togglePrevious(false);
          toggleNext((buttonCount > displayCount));
          
          togglePageButton();
          
          
          function togglePrevious(status) {
            if(status){
              element.find('ul li:first').removeClass('disabled');
            }else{
              element.find('ul li:first').addClass('disabled');
            }
          }
                    
          function toggleNext(status) {
            if(status) {
              element.find('ul li:last').removeClass('disabled');
            }else{
              element.find('ul li:last').addClass('disabled');
            }
          }
          
          function buttonClickHandler(e) {
            ctrl.page = $(e.target).attr('page');
            togglePageButton();
            ctrl.retrieve(); 
                                      
            if(tc.canIncrement()) {
              toggleNext(true);
            }else {
              toggleNext(false);
            }
            
            if(tc.canDecrement()) {
              togglePrevious(true);
            }else{
              togglePrevious(false);
            }
          }  
          
          function togglePageButton() {
             element.find('li[tag="pagination-button"]').removeClass('active');
             element.find('li[tag="pagination-button"] a[page="' + ctrl.page + '"]').parent().addClass('active');
          }          
         
          function previous() {
            if(tc.canDecrement()) {
              tc.decrement();
              drawButtons(tc.getTime());
              element.find('li[tag="pagination-button"] a[page="' + ctrl.page + '"]').trigger('click');      
            }
          }      
          
          function next() {
            if(tc.canIncrement()) {    
              tc.increment();
              drawButtons(tc.getTime());     
              element.find('li[tag="pagination-button"] a[page="' + ctrl.page + '"]').trigger('click');
              
            }
          }
        }
      }); 
    }
  }
  
})();