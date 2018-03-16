import expect from 'expect';
import Queues from './Queues';
import React from 'react';
import ReactTestUtils from 'react-addons-test-utils';

describe('Queues', () => {
  it('gets queued count', () => {
    let r = ReactTestUtils.createRenderer();
    r.render(<Queues />);
    let queues = r.getMountedInstance();
    expect(queues.state.queues.length).toEqual(0);

    queues.setState({
      queues: [
        {job_name: 'test', count: 1, latency: 0},
        {job_name: 'test2', count: 2, latency: 0}
      ]
    });

    expect(queues.state.queues.length).toEqual(2);
    expect(queues.queuedCount).toEqual(3);
  });
});
