import { delUrlParam } from "./utils";

describe('functions in utils.ts should work', () => {

  it('function delUrlParam() should work', () => {
    expect(delUrlParam).toBeTruthy();
    expect(delUrlParam('http://test.com?param1=a&param2=b&param3=c', 'param2'))
      .toEqual('http://test.com?param1=a&param3=c');
    expect(delUrlParam('http://test.com', 'param2')).toEqual('http://test.com');
    expect(delUrlParam('http://test.com?param2', 'param2')).toEqual('http://test.com');
  });
});
