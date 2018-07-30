import './TestSetup';
import expect from 'expect';
import DeadJobs from './DeadJobs';
import React from 'react';
import { mount } from 'enzyme';

describe('DeadJobs', () => {
  it('shows dead jobs', () => {
    let deadJobs = mount(<DeadJobs />);

    expect(deadJobs.state().selected.length).toEqual(0);
    expect(deadJobs.state().jobs.length).toEqual(0);

    deadJobs.setState({
      count: 2,
      jobs: [
        {id: 1, name: 'test', args: {}, t: 1467760821, err: 'err1'},
        {id: 2, name: 'test2', args: {}, t: 1467760822, err: 'err2'}
      ]
    });

    expect(deadJobs.state().selected.length).toEqual(0);
    expect(deadJobs.state().jobs.length).toEqual(2);

    let checkbox = deadJobs.find('input');
    expect(checkbox.length).toEqual(3);
    expect(checkbox.at(0).props().checked).toEqual(false);
    expect(checkbox.at(1).props().checked).toEqual(false);
    expect(checkbox.at(2).props().checked).toEqual(false);

    checkbox.at(0).simulate('change');
    checkbox = deadJobs.find('input');
    expect(checkbox.length).toEqual(3);
    expect(checkbox.at(0).props().checked).toEqual(true);
    expect(checkbox.at(1).props().checked).toEqual(true);
    expect(checkbox.at(2).props().checked).toEqual(true);

    checkbox.at(1).simulate('change');
    checkbox = deadJobs.find('input');
    expect(checkbox.length).toEqual(3);
    expect(checkbox.at(0).props().checked).toEqual(true);
    expect(checkbox.at(1).props().checked).toEqual(false);
    expect(checkbox.at(2).props().checked).toEqual(true);

    checkbox.at(1).simulate('change');
    checkbox = deadJobs.find('input');
    expect(checkbox.length).toEqual(3);
    expect(checkbox.at(0).props().checked).toEqual(true);
    expect(checkbox.at(1).props().checked).toEqual(true);
    expect(checkbox.at(2).props().checked).toEqual(true);

    let button = deadJobs.find('button');
    expect(button.length).toEqual(4);
    button.at(0).simulate('click');
    button.at(1).simulate('click');
    button.at(2).simulate('click');
    button.at(3).simulate('click');

    checkbox.at(0).simulate('change');

    checkbox = deadJobs.find('input');
    expect(checkbox.length).toEqual(3);
    expect(checkbox.at(0).props().checked).toEqual(false);
    expect(checkbox.at(1).props().checked).toEqual(false);
    expect(checkbox.at(2).props().checked).toEqual(false);
  });

  it('has pages', () => {
    let deadJobs = mount(<DeadJobs />);

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
    deadJobs.setState({
      count: 21,
      jobs: genJob(21)
    });

    expect(deadJobs.state().jobs.length).toEqual(21);
    expect(deadJobs.state().page).toEqual(1);

    let pageList = deadJobs.find('PageList');
    expect(pageList.length).toEqual(1);

    pageList.at(0).props().jumpTo(2)();
    expect(deadJobs.state().page).toEqual(2);
  });
});
