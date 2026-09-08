// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
