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
        'displayCount': '@'
      },
      'link': link,
      'controller': PaginatorController,
      'controllerAs': 'vm',
      'bindToController': true
    };
    return directive;
    
    function link(scope, element, attrs, ctrl) {
      
      scope.$watch('vm.page', function(current) {
        if(current) { 
          ctrl.page = current;
          togglePageButton();
        }
      });
      
      var tc;
                        
      scope.$watch('vm.totalCount', function(current) {
        if(current) {
          var totalCount   = current;   
                                              
          element.find('ul li:first a').off('click');
          element.find('ul li:last a').off('click');
                    
          tc = new TimeCounter();
          
          console.log('Total Count:' + totalCount + ', Page Size:' + ctrl.pageSize + ', Display Count:' + ctrl.displayCount + ', Page:' + ctrl.page);

          ctrl.buttonCount = Math.ceil(totalCount / ctrl.pageSize);
                                
          if(ctrl.buttonCount <= ctrl.displayCount) {
            tc.setMaximum(1);
          }else{
            tc.setMaximum(Math.ceil(ctrl.buttonCount / ctrl.displayCount));
          }
                   
          element.find('ul li:first a').on('click', previous);          
          element.find('ul li:last a').on('click', next);          
          
          drawButtons(tc.getTime());    

          togglePrevious(tc.canDecrement());
          toggleNext(tc.canIncrement());
          
          togglePageButton();
          
        }
      }); 

      var TimeCounter = function() {
        this.time = 0;
        this.minimum = 0;
        this.maximum = 0;
      };
      
      TimeCounter.prototype.setMaximum = function(maximum) {
        this.maximum = maximum;
      };
      
      TimeCounter.prototype.increment = function() {
        if(this.time < this.maximum) {
          ++this.time;
          if((ctrl.page % ctrl.displayCount) != 0) {
            ctrl.page = this.time * ctrl.displayCount;
          }
          ++ctrl.page;
        }
        scope.$apply();
      };
      
      TimeCounter.prototype.canIncrement = function() {
        if(this.time + 1 < this.maximum) {
          return true;
        }
        return false;
      };
      
      TimeCounter.prototype.decrement = function() {
        if(this.time > this.minimum) {         
          if(this.time === 0) {
            ctrl.page = ctrl.displayCount;
          }else if((ctrl.page % ctrl.displayCount) != 0) {
            ctrl.page =  this.time * ctrl.displayCount;
          }
          --this.time;
          --ctrl.page;
        }
        scope.$apply();
      };
      
      TimeCounter.prototype.canDecrement = function() {
        if(this.time > this.minimum) {
          return true;
        }
        return false;
      };
      
      TimeCounter.prototype.getTime = function() {
        return this.time;
      };
                 
      function drawButtons(time) {
        element.find('li[tag="pagination-button"]').remove();
        var buttons = [];
        for(var i = 1; i <= ctrl.displayCount; i++) {
          var displayNumber = ctrl.displayCount * time + i;
          if(displayNumber <= ctrl.buttonCount) {
            buttons.push('<li tag="pagination-button"><a href="javascript:void(0)" page="' + displayNumber + '">' + displayNumber + '<span class="sr-only"></span></a></li>');
          }
        }
        $(buttons.join(''))
          .insertAfter(element.find('ul li:eq(0)')).end()
          .on('click', buttonClickHandler); 
      }
      
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
        togglePrevious(tc.canDecrement());
        toggleNext(tc.canIncrement());
        scope.$apply();
      }  
          
      function togglePageButton() {
        element.find('li[tag="pagination-button"]').removeClass('active');
        element.find('li[tag="pagination-button"] a[page="' + ctrl.page + '"]').parent().addClass('active');
      }          
         
      function previous() {
        if(tc.canDecrement()) {
          tc.decrement();
          drawButtons(tc.getTime());
          togglePageButton();
          togglePrevious(tc.canDecrement());
          toggleNext(tc.canIncrement());
        }
        scope.$apply(); 
      }      
          
      function next() {
        if(tc.canIncrement()) {    
          tc.increment();
          drawButtons(tc.getTime());     
          togglePageButton();
          togglePrevious(tc.canDecrement());
          toggleNext(tc.canIncrement());
        }
        scope.$apply();
      }
    }
  }
  
})();