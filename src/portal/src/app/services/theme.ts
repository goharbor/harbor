export enum StyleMode {
    DARK = 'DARK',
    LIGHT = 'LIGHT',
}

export const HAS_STYLE_MODE: string = 'styleModeLocal';

export interface ThemeInterface {
    showStyle: string;
    mode: string;
    text: string;
    currentFileName: string;
    toggleFileName: string;
}

export interface CustomStyle {
    headerBgColor: {
        darkMode: string;
        lightMode: string;
    };
    loginBgImg: string;
    loginTitle: string;
    product: {
        name: string;
        logo: string;
        introduction: string;
    };
}

export const THEME_ARRAY: ThemeInterface[] = [
    {
        showStyle: StyleMode.DARK,
        mode: StyleMode.LIGHT,
        text: 'APP_TITLE.THEME_LIGHT_TEXT',
        currentFileName: 'dark-theme.css',
        toggleFileName: 'light-theme.css',
    },
    {
        showStyle: StyleMode.LIGHT,
        mode: StyleMode.DARK, // show button icon
        text: 'APP_TITLE.THEME_DARK_TEXT', // show button text
        currentFileName: 'light-theme.css', // loaded current theme file name
        toggleFileName: 'dark-theme.css', // to toggle theme file name
    },
];
