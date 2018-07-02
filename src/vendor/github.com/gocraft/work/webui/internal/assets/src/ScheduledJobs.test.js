import './TestSetup';
import expect from 'expect';
import ScheduledJobs from './ScheduledJobs';
import React from 'react';
import { mount } from 'enzyme';

describe('ScheduledJobs', () => {
  it('shows jobs', () => {
    let scheduledJobs = mount(<ScheduledJobs />);

    expect(scheduledJobs.state().jobs.length).toEqual(0);

    scheduledJobs.setState({
      count: 2,
      jobs: [
        {id: 1, name: 'test', args: {}, run_at: 1467760821, err: 'err1'},
        {id: 2, name: 'test2', args: {}, run_at: 1467760822, err: 'err2'}
      ]
    });

    expect(scheduledJobs.state().jobs.length).toEqual(2);
  });

  it('has pages', () => {
    let scheduledJobs = mount(<ScheduledJobs />);

    let genJob = (n) => {
      let job = [];
      for (let i = 1; i <= n; i++) {
        job.push({
          id: i,
          name: 'test',
          args: {},
          run_at: 1467760821,
          err: 'err',
        });
      }
      return job;
    };
    scheduledJobs.setState({
      count: 21,
      jobs: genJob(21)
    });

    expect(scheduledJobs.state().jobs.length).toEqual(21);
    expect(scheduledJobs.state().page).toEqual(1);

    let pageList = scheduledJobs.find('PageList');
    expect(pageList.length).toEqual(1);

    pageList.at(0).props().jumpTo(2)();
    expect(scheduledJobs.state().page).toEqual(2);
  });
});
