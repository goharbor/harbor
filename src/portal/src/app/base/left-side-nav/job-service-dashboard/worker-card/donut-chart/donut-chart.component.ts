import { Component, Input } from '@angular/core';
import { HAS_STYLE_MODE, StyleMode } from '../../../../../services/theme';

// this donut chart is created based on css border-radius and transform: rotate()
// as we have two themes, so need to set different colors for different theme
// for dark theme: .center bg-color #21333b; denominator #405C6B; numerator #49aeda
// for light theme: .center bg-color #fff; denominator #ccc; numerator #0072a3

enum DarkColors {
    CENTER_BG_COLOR = '#21333b',
    DEN_COLOR = '#405C6B',
    NUM_COLOR = '#49aeda',
}

enum LightColors {
    CENTER_BG_COLOR = '#fff',
    DEN_COLOR = '#ccc',
    NUM_COLOR = '#0072a3',
}

@Component({
    selector: 'app-donut-chart',
    templateUrl: './donut-chart.component.html',
    styleUrls: ['./donut-chart.component.scss'],
})
export class DonutChartComponent {
    @Input()
    denominator: number;

    @Input()
    numerator: number;

    isDarkTheme(): boolean {
        return localStorage?.getItem(HAS_STYLE_MODE) === StyleMode.DARK;
    }

    getSmallBGColor(): string {
        if (this.isDarkTheme()) {
            return DarkColors.CENTER_BG_COLOR;
        }
        return LightColors.CENTER_BG_COLOR;
    }

    getBigBGColor(): string {
        if (this.isDarkTheme()) {
            return DarkColors.NUM_COLOR;
        }
        return LightColors.NUM_COLOR;
    }

    getLeftBGColor() {
        if (this.isDarkTheme()) {
            return DarkColors.DEN_COLOR;
        }
        return LightColors.DEN_COLOR;
    }

    getRightBGColor(): string {
        if (this.getDegree() > 180) {
            if (this.isDarkTheme()) {
                return DarkColors.NUM_COLOR;
            }
            return LightColors.NUM_COLOR;
        }
        if (this.isDarkTheme()) {
            return DarkColors.DEN_COLOR;
        }
        return LightColors.DEN_COLOR;
    }

    getDegree(): number {
        if (this.numerator && this.denominator) {
            return (this.numerator / this.denominator) * 360;
        }
        return 0;
    }

    getRightRotate() {
        if (this.getDegree() > 0 && this.getDegree() <= 180) {
            return `rotate(${this.getDegree()}deg)`;
        }
        return `rotate(0)`;
    }

    getLeftRotate() {
        if (this.getDegree() > 180) {
            return `rotate(${this.getDegree() - 180}deg)`;
        }
        return `rotate(0)`;
    }
}
