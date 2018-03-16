import expect from 'expect';
import UnixTime from './UnixTime';
import React from 'react';
import ReactTestUtils from 'react-addons-test-utils';

describe('UnixTime', () => {
  it('formats human-readable time string', () => {
    let r = ReactTestUtils.createRenderer();
    r.render(<UnixTime ts={1467753603} />);
    let output = r.getRenderOutput();

    expect(output.type).toEqual('time');
    expect(output.props.children).toEqual('2016/07/05 21:20:03');
    expect(output.props.dateTime).toEqual('2016-07-05T21:20:03.000Z');
  });
});
