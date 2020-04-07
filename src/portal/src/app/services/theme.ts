export interface ThemeInterface {
    showStyle: string;
    mode: string;
    text: string;
    currentFileName: string;
    toggleFileName: string;
}

export const THEME_ARRAY: ThemeInterface[] = [
    {
        showStyle: "DARK",
        mode: "LIGHT",
        text: "APP_TITLE.THEME_LIGHT_TEXT",
        currentFileName: "dark-theme.css",
        toggleFileName: "light-theme.css",
    },
    {
        showStyle: "LIGHT",
        mode: "DARK", // show button icon
        text: "APP_TITLE.THEME_DARK_TEXT", // show button text
        currentFileName: "light-theme.css", // loaded current theme file name
        toggleFileName: "dark-theme.css", // to toggle theme file name
    }
];
