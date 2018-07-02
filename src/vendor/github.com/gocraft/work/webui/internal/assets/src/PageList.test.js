import './TestSetup';
import expect from 'expect';
import PageList from './PageList';
import React from 'react';
import { mount } from 'enzyme';

describe('PageList', () => {
  it('lists pages', () => {
    let assertPage = (n, expected) => {
      let pageList = mount(<PageList page={n} perPage={2} totalCount={13} jumpTo={() => () => {}} />);
      let ul = pageList.find('ul');

      expect(ul.props().children.map((el) => {
        expect(el.type).toEqual('li');
        return el.props.children.props.children;
      })).toEqual(expected);
    };

    assertPage(1, [1, 2, '..', 7]);
    assertPage(2, [1, 2, 3, '..', 7]);
    assertPage(3, [1, 2, 3, 4, '..', 7]);
    assertPage(4, [1, '..', 3, 4, 5, '..', 7]);
    assertPage(5, [1, '..', 4, 5, 6, 7]);
    assertPage(6, [1, '..', 5, 6, 7]);
    assertPage(7, [1, '..', 6, 7]);
  });

  it('renders nothing if there is nothing', () => {
    let pageList = mount(<PageList page={1} perPage={2} totalCount={0} jumpTo={() => () => {}} />);

    expect(pageList.html()).toEqual(null);
  });
});
