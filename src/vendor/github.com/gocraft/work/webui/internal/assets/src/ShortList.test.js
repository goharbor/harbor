import expect from 'expect';
import ShortList from './ShortList';
import React from 'react';
import ReactTestUtils from 'react-addons-test-utils';

describe('ShortList', () => {
  it('lists items', () => {
    let r = ReactTestUtils.createRenderer();
    r.render(<ShortList item={['1', '2', '3', '4']} />);
    let output = r.getRenderOutput();

    expect(output.type).toEqual('ul');
    output.props.children.map((el, i) => {
      expect(el.type).toEqual('li');
      if (i < 3) {
        expect(el.props.children).toEqual(i+1);
      } else {
        expect(el.props.children).toEqual([i-2, ' more']);
      }
    });
  });
});
