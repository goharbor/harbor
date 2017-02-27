import { Directive, ElementRef, HostListener } from '@angular/core';

export const customColor = 'blue';
export const customFontColor = 'white';

@Directive({
  selector: '[custom-highlight]'
})
export class CustomHighlightDirective {
  constructor(private el: ElementRef) {}

  @HostListener('mouseenter')
  onMouseEnter(): void {
    this.el.nativeElement.style.backgroundColor = customColor;
    this.el.nativeElement.style.color = customFontColor;
  }

  @HostListener('mouseout')
  onMouseOut(): void {
    this.el.nativeElement.style.backgroundColor = null;
    this.el.nativeElement.style.color = null;
  }
}