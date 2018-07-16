import './TestSetup';
import expect from 'expect';
import ShortList from './ShortList';
import React from 'react';
import { mount } from 'enzyme';

describe('ShortList', () => {
  it('lists items', () => {
    let shortList = mount(<ShortList item={['1', '2', '3', '4']} />);
    let ul = shortList.find('ul');

    ul.props().children.map((el, i) => {
      expect(el.type).toEqual('li');
      if (i < 3) {
        expect(el.props.children).toEqual(String(i+1));
      } else {
        expect(el.props.children).toEqual([i-2, ' more']);
      }
    });
  });
});
