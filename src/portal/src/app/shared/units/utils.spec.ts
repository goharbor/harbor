import { isSameArrayValue, isSameObject } from "./utils";

describe('functions in utils.ts should work', () => {
  it('function isSameArrayValue() should work', () => {
    expect(isSameArrayValue).toBeTruthy();
    expect(isSameArrayValue(null, null)).toBeFalsy();
    expect(isSameArrayValue([], null)).toBeFalsy();
    expect(isSameArrayValue([1, 2, 3], [3 , 2, 1])).toBeTruthy();
    expect(isSameArrayValue([{a: 1, c: 2}, true], [true, {c: 2, a: 1, d: null}])).toBeTruthy();
  });

  it('function isSameObject() should work', () => {
    expect(isSameObject).toBeTruthy();
    expect(isSameObject(null, null)).toBeTruthy();
    expect(isSameObject({}, null)).toBeFalsy();
    expect(isSameObject(null, {})).toBeFalsy();
    expect(isSameObject([], null)).toBeFalsy();
    expect(isSameObject(null, [])).toBeFalsy();
    expect(isSameObject({a: 1, b: true}, {a: 1})).toBeFalsy();
    expect(isSameObject({a: 1, b: false}, {a: 1})).toBeFalsy();
    expect(isSameObject({a: [1, 2, 3], b: null}, {a: [3, 2, 1]})).toBeTruthy();
    expect(isSameObject({a: {a: 1 , b: 2}, b: null}, {a: {b: 2, a: 1}})).toBeTruthy();
    expect(isSameObject([1, 2, 3], [3 , 2, 1])).toBeFalsy();
  });
});
