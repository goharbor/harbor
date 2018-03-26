import expect from 'expect';
import Processes from './Processes';
import React from 'react';
import ReactTestUtils from 'react-addons-test-utils';
import { findAllByTag } from './TestUtils';

describe('Processes', () => {
  it('shows workers', () => {
    let r = ReactTestUtils.createRenderer();
    r.render(<Processes />);
    let processes = r.getMountedInstance();

    expect(processes.state.busyWorker.length).toEqual(0);
    expect(processes.state.workerPool.length).toEqual(0);

    processes.setState({
      busyWorker: [
        {
          worker_id: '2',
          job_name: 'job1',
          started_at: 1467753603,
          checkin_at: 1467753603,
          checkin: '123',
          args_json: '{}'
        }
      ],
      workerPool: [
        {
          worker_pool_id: '1',
          started_at: 1467753603,
          heartbeat_at: 1467753603,
          job_names: ['job1', 'job2', 'job3', 'job4'],
          concurrency: 10,
          host: 'web51',
          pid: 123,
          worker_ids: [
            '1', '2', '3'
          ]
        }
      ]
    });

    expect(processes.state.busyWorker.length).toEqual(1);
    expect(processes.state.workerPool.length).toEqual(1);
    expect(processes.workerCount).toEqual(3);

    const expectedBusyWorker = [ { args_json: '{}', checkin: '123', checkin_at: 1467753603, job_name: 'job1', started_at: 1467753603, worker_id: '2' } ];

    let output = r.getRenderOutput();
    let busyWorkers = findAllByTag(output, 'BusyWorkers');
    expect(busyWorkers.length).toEqual(1);
    expect(busyWorkers[0].props.worker).toEqual(expectedBusyWorker);
    expect(processes.getBusyPoolWorker(processes.state.workerPool[0])).toEqual(expectedBusyWorker);
  });
});
