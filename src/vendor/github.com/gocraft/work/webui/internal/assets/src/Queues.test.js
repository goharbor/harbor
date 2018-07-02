import './TestSetup';
import expect from 'expect';
import Queues from './Queues';
import React from 'react';
import { mount } from 'enzyme';

describe('Queues', () => {
  it('gets queued count', () => {
    let queues = mount(<Queues />);
    expect(queues.state().queues.length).toEqual(0);

    queues.setState({
      queues: [
        {job_name: 'test', count: 1, latency: 0},
        {job_name: 'test2', count: 2, latency: 0}
      ]
    });

    expect(queues.state().queues.length).toEqual(2);
    expect(queues.instance().queuedCount).toEqual(3);
  });
});
