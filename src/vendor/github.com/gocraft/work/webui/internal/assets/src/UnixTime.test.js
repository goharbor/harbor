import './TestSetup';
import expect from 'expect';
import UnixTime from './UnixTime';
import React from 'react';
import { mount } from 'enzyme';

describe('UnixTime', () => {
  it('formats human-readable time string', () => {
    let output = mount(<UnixTime ts={1467753603} />);

    let time = output.find('time');
    expect(time.props().dateTime).toEqual('2016-07-05T21:20:03.000Z');
    expect(time.text()).toEqual('2016/07/05 21:20:03');
  });
});
