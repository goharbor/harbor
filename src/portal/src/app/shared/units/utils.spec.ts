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
import {
    DEFAULT_PAGE_SIZE,
    delUrlParam,
    durationStr,
    getHiddenArrayFromLocalStorage,
    getPageSizeFromLocalStorage,
    getQueryString,
    getSizeNumber,
    getSizeUnit,
    getSortingString,
    isSameArrayValue,
    isSameObject,
    setHiddenArrayToLocalStorage,
    setPageSizeToLocalStorage,
    calculateLuminance,
    getTextColorForBackground,
    isValidHexColor,
} from './utils';
import { ClrDatagridStateInterface } from '@clr/angular';
import { QuotaUnit } from '../entities/shared.const';

describe('functions in utils.ts should work', () => {
    it('function isSameArrayValue() should work', () => {
        expect(isSameArrayValue).toBeTruthy();
        expect(isSameArrayValue(null, null)).toBeFalsy();
        expect(isSameArrayValue([], null)).toBeFalsy();
        expect(isSameArrayValue([1, 2, 3], [3, 2, 1])).toBeTruthy();
        expect(
            isSameArrayValue(
                [{ a: 1, c: 2 }, true],
                [true, { c: 2, a: 1, d: null }]
            )
        ).toBeTruthy();
    });

    it('function isSameObject() should work', () => {
        expect(isSameObject).toBeTruthy();
        expect(isSameObject(null, null)).toBeTruthy();
        expect(isSameObject({}, null)).toBeFalsy();
        expect(isSameObject(null, {})).toBeFalsy();
        expect(isSameObject([], null)).toBeFalsy();
        expect(isSameObject(null, [])).toBeFalsy();
        expect(isSameObject({ a: 1, b: true }, { a: 1 })).toBeFalsy();
        expect(isSameObject({ a: 1, b: false }, { a: 1 })).toBeFalsy();
        expect(
            isSameObject({ a: [1, 2, 3], b: null }, { a: [3, 2, 1] })
        ).toBeTruthy();
        expect(
            isSameObject({ a: { a: 1, b: 2 }, b: null }, { a: { b: 2, a: 1 } })
        ).toBeTruthy();
        expect(isSameObject([1, 2, 3], [3, 2, 1])).toBeFalsy();
    });

    it('function delUrlParam() should work', () => {
        expect(delUrlParam).toBeTruthy();
        expect(
            delUrlParam('http://test.com?param1=a&param2=b&param3=c', 'param2')
        ).toEqual('http://test.com?param1=a&param3=c');
        expect(delUrlParam('http://test.com', 'param2')).toEqual(
            'http://test.com'
        );
        expect(delUrlParam('http://test.com?param2', 'param2')).toEqual(
            'http://test.com'
        );
    });

    it('function getSortingString() should work', () => {
        expect(getSortingString).toBeTruthy();
        const state: ClrDatagridStateInterface = {
            sort: {
                by: 'name',
                reverse: true,
            },
        };
        expect(getSortingString(state)).toEqual('-name');
    });

    it('function getQueryString() should work', () => {
        expect(getQueryString).toBeTruthy();
        const state: ClrDatagridStateInterface = {
            filters: [
                { property: 'name', value: 'test' },
                { property: 'url', value: 'http://test.com' },
            ],
        };
        expect(getQueryString(state)).toEqual(
            encodeURIComponent('name=~test,url=~http://test.com')
        );
    });

    it('function getSizeNumber() should work', () => {
        expect(getSizeNumber).toBeTruthy();
        expect(getSizeNumber(4564)).toEqual('4.46');
        expect(getSizeNumber(10)).toEqual(10);
        expect(getSizeNumber(456400)).toEqual('445.70');
        expect(getSizeNumber(45640000)).toEqual('43.53');
        expect(getSizeNumber(4564000000000)).toEqual('4.15');
    });

    it('function getSizeUnit() should work', () => {
        expect(getSizeUnit).toBeTruthy();
        expect(getSizeUnit(4564)).toEqual(QuotaUnit.KB);
        expect(getSizeUnit(10)).toEqual(QuotaUnit.BIT);
        expect(getSizeUnit(4564000)).toEqual(QuotaUnit.MB);
        expect(getSizeUnit(4564000000)).toEqual(QuotaUnit.GB);
        expect(getSizeUnit(4564000000000)).toEqual(QuotaUnit.TB);
    });

    it('functions getPageSizeFromLocalStorage() and setPageSizeToLocalStorage() should work', () => {
        let store = {};
        spyOn(localStorage, 'getItem').and.callFake(key => {
            return store[key];
        });
        spyOn(localStorage, 'setItem').and.callFake((key, value) => {
            return (store[key] = value + '');
        });
        spyOn(localStorage, 'clear').and.callFake(() => {
            store = {};
        });
        expect(getPageSizeFromLocalStorage(null)).toEqual(DEFAULT_PAGE_SIZE);
        expect(getPageSizeFromLocalStorage('test', 99)).toEqual(99);
        expect(getPageSizeFromLocalStorage('')).toEqual(DEFAULT_PAGE_SIZE);
        setPageSizeToLocalStorage('test1', null);
        expect(getPageSizeFromLocalStorage('test1')).toEqual(DEFAULT_PAGE_SIZE);
        setPageSizeToLocalStorage('test1', 10);
        expect(getPageSizeFromLocalStorage('test1')).toEqual(10);
    });

    it('functions durationStr(distance: number) should work', () => {
        expect(durationStr(11)).toEqual('0');
        expect(durationStr(1111)).toEqual('1sec');
        expect(durationStr(61111)).toEqual('1min 1sec');
        expect(durationStr(3661111)).toEqual('1hrs 1min 1sec');
    });

    it('functions getHiddenArrayFromLocalStorage() and setHiddenArrayToLocalStorage() should work', () => {
        let store = {};
        spyOn(localStorage, 'getItem').and.callFake(key => {
            return store[key];
        });
        spyOn(localStorage, 'setItem').and.callFake((key, value) => {
            return (store[key] = value + '');
        });
        spyOn(localStorage, 'clear').and.callFake(() => {
            store = {};
        });
        expect(getHiddenArrayFromLocalStorage(null, [])).toEqual([]);
        expect(getHiddenArrayFromLocalStorage('test', [true])).toEqual([true]);
        expect(getHiddenArrayFromLocalStorage('test1', [])).toEqual([]);
        setHiddenArrayToLocalStorage('test1', [false, false, false]);
        expect(getHiddenArrayFromLocalStorage('test1', [false])).toEqual([
            false,
            false,
            false,
        ]);
        setHiddenArrayToLocalStorage('test1', [true, true]);
        expect(getHiddenArrayFromLocalStorage('test1', [false])).toEqual([
            true,
            true,
        ]);
    });

    it('function calculateLuminance() should work', () => {
        expect(calculateLuminance).toBeTruthy();
        // Test black color (low luminance)
        expect(calculateLuminance('#000000')).toBeLessThan(0.1);
        // Test white color (high luminance)
        expect(calculateLuminance('#FFFFFF')).toBeGreaterThan(0.9);
        // Test without # prefix
        expect(calculateLuminance('FFFFFF')).toBeGreaterThan(0.9);
        // Test a mid-range color
        const midLuminance = calculateLuminance('#808080');
        expect(midLuminance).toBeGreaterThan(0.1);
        expect(midLuminance).toBeLessThan(0.9);
    });

    it('function getTextColorForBackground() should work', () => {
        expect(getTextColorForBackground).toBeTruthy();
        // Dark backgrounds should have white text
        expect(getTextColorForBackground('#000000')).toEqual('white');
        expect(getTextColorForBackground('#0065AB')).toEqual('white');
        // Light backgrounds should have black text
        expect(getTextColorForBackground('#FFFFFF')).toEqual('black');
        expect(getTextColorForBackground('#FFDC0B')).toEqual('black');
        // Test without # prefix
        expect(getTextColorForBackground('000000')).toEqual('white');
        expect(getTextColorForBackground('FFFFFF')).toEqual('black');
    });

    it('function isValidHexColor() should work', () => {
        expect(isValidHexColor).toBeTruthy();
        // Valid 6-digit hex codes
        expect(isValidHexColor('#FFFFFF')).toBeTruthy();
        expect(isValidHexColor('#000000')).toBeTruthy();
        expect(isValidHexColor('#abc123')).toBeTruthy();
        // Valid 3-digit hex codes
        expect(isValidHexColor('#FFF')).toBeTruthy();
        expect(isValidHexColor('#000')).toBeTruthy();
        expect(isValidHexColor('#a1b')).toBeTruthy();
        // Valid without # prefix
        expect(isValidHexColor('FFFFFF')).toBeTruthy();
        expect(isValidHexColor('FFF')).toBeTruthy();
        // Invalid hex codes
        expect(isValidHexColor('#GGGGGG')).toBeFalsy();
        expect(isValidHexColor('#12345')).toBeFalsy();
        expect(isValidHexColor('invalid')).toBeFalsy();
        expect(isValidHexColor('')).toBeFalsy();
    });
});
