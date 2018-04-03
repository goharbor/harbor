import expect from 'expect';
import DeadJobs from './DeadJobs';
import React from 'react';
import ReactTestUtils from 'react-addons-test-utils';
import { findAllByTag } from './TestUtils';

describe('DeadJobs', () => {
  it('shows dead jobs', () => {
    let r = ReactTestUtils.createRenderer();
    r.render(<DeadJobs />);
    let deadJobs = r.getMountedInstance();

    expect(deadJobs.state.selected.length).toEqual(0);
    expect(deadJobs.state.jobs.length).toEqual(0);

    deadJobs.setState({
      count: 2,
      jobs: [
        {id: 1, name: 'test', args: {}, t: 1467760821, err: 'err1'},
        {id: 2, name: 'test2', args: {}, t: 1467760822, err: 'err2'}
      ]
    });

    expect(deadJobs.state.selected.length).toEqual(0);
    expect(deadJobs.state.jobs.length).toEqual(2);

    let output = r.getRenderOutput();
    let checkbox = findAllByTag(output, 'input');
    expect(checkbox.length).toEqual(3);

    expect(checkbox[0].props.checked).toEqual(false);
    expect(checkbox[1].props.checked).toEqual(false);
    expect(checkbox[2].props.checked).toEqual(false);

    checkbox[0].props.onChange();

    output = r.getRenderOutput();
    checkbox = findAllByTag(output, 'input');
    expect(checkbox.length).toEqual(3);
    expect(checkbox[0].props.checked).toEqual(true);
    expect(checkbox[1].props.checked).toEqual(true);
    expect(checkbox[2].props.checked).toEqual(true);

    checkbox[1].props.onChange();

    output = r.getRenderOutput();
    checkbox = findAllByTag(output, 'input');
    expect(checkbox.length).toEqual(3);
    expect(checkbox[0].props.checked).toEqual(true);
    expect(checkbox[1].props.checked).toEqual(false);
    expect(checkbox[2].props.checked).toEqual(true);

    checkbox[1].props.onChange();
    output = r.getRenderOutput();
    checkbox = findAllByTag(output, 'input');
    expect(checkbox.length).toEqual(3);
    expect(checkbox[0].props.checked).toEqual(true);
    expect(checkbox[1].props.checked).toEqual(true);
    expect(checkbox[2].props.checked).toEqual(true);

    let button = findAllByTag(output, 'button');
    expect(button.length).toEqual(4);
    button[0].props.onClick();
    button[1].props.onClick();
    button[2].props.onClick();
    button[3].props.onClick();

    checkbox[0].props.onChange();

    output = r.getRenderOutput();
    checkbox = findAllByTag(output, 'input');
    expect(checkbox.length).toEqual(3);
    expect(checkbox[0].props.checked).toEqual(false);
    expect(checkbox[1].props.checked).toEqual(false);
    expect(checkbox[2].props.checked).toEqual(false);
  });

  it('has pages', () => {
    let r = ReactTestUtils.createRenderer();
    r.render(<DeadJobs />);
    let deadJobs = r.getMountedInstance();

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

    expect(deadJobs.state.jobs.length).toEqual(21);
    expect(deadJobs.state.page).toEqual(1);

    let output = r.getRenderOutput();
    let pageList = findAllByTag(output, 'PageList');
    expect(pageList.length).toEqual(1);

    pageList[0].props.jumpTo(2)();
    expect(deadJobs.state.page).toEqual(2);
  });
});
