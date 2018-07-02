import './TestSetup';
import expect from 'expect';
import RetryJobs from './RetryJobs';
import React from 'react';
import { mount } from 'enzyme';

describe('RetryJobs', () => {
  it('shows jobs', () => {
    let retryJobs = mount(<RetryJobs />);

    expect(retryJobs.state().jobs.length).toEqual(0);

    retryJobs.setState({
      count: 2,
      jobs: [
        {id: 1, name: 'test', args: {}, t: 1467760821, err: 'err1'},
        {id: 2, name: 'test2', args: {}, t: 1467760822, err: 'err2'}
      ]
    });

    expect(retryJobs.state().jobs.length).toEqual(2);
  });

  it('has pages', () => {
    let retryJobs = mount(<RetryJobs />);

    let genJob = (n) => {
      let job = [];
      for (let i = 1; i <= n; i++) {
        job.push({
          id: i,
          name: 'test',
          args: {},
          t: 1467760821,
          err: 'err',
        });
      }
      return job;
    };
    retryJobs.setState({
      count: 21,
      jobs: genJob(21)
    });

    expect(retryJobs.state().jobs.length).toEqual(21);
    expect(retryJobs.state().page).toEqual(1);

    let pageList = retryJobs.find('PageList');
    expect(pageList.length).toEqual(1);

    pageList.at(0).props().jumpTo(2)();
    expect(retryJobs.state().page).toEqual(2);
  });
});
