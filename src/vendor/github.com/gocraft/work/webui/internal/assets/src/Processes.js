import React from 'react';
import PropTypes from 'prop-types';
import UnixTime from './UnixTime';
import ShortList from './ShortList';
import styles from './bootstrap.min.css';
import cx from './cx';

class BusyWorkers extends React.Component {
  static propTypes = {
    worker: PropTypes.arrayOf(PropTypes.object).isRequired,
  }

  render() {
    return (
      <div className={styles.tableResponsive}>
        <table className={styles.table}>
          <tbody>
            <tr>
              <th>Name</th>
              <th>Arguments</th>
              <th>Started At</th>
              <th>Check-in At</th>
              <th>Check-in</th>
            </tr>
            {
              this.props.worker.map((worker) => {
                return (
                  <tr key={worker.worker_id}>
                    <td>{worker.job_name}</td>
                    <td>{worker.args_json}</td>
                    <td><UnixTime ts={worker.started_at}/></td>
                    <td><UnixTime ts={worker.checkin_at}/></td>
                    <td>{worker.checkin}</td>
                  </tr>
                );
              })
            }
          </tbody>
        </table>
      </div>
    );
  }
}

export default class Processes extends React.Component {
  static propTypes = {
    busyWorkerURL: PropTypes.string,
    workerPoolURL: PropTypes.string,
  }

  state = {
    busyWorker: [],
    workerPool: []
  }

  componentWillMount() {
    if (this.props.busyWorkerURL) {
      fetch(this.props.busyWorkerURL).
        then((resp) => resp.json()).
        then((data) => {
          if (data) {
            this.setState({
              busyWorker: data
            });
          }
        });
    }
    if (this.props.workerPoolURL) {
      fetch(this.props.workerPoolURL).
        then((resp) => resp.json()).
        then((data) => {
          let workers = [];
          data.map((worker) => {
            if (worker.host != '') {
              workers.push(worker);
            }
          });
          this.setState({
            workerPool: workers
          });
        });
    }
  }

  get workerCount() {
    let count = 0;
    this.state.workerPool.map((pool) => {
      count += pool.worker_ids.length;
    });
    return count;
  }

  getBusyPoolWorker(pool) {
    let workers = [];
    this.state.busyWorker.map((worker) => {
      if (pool.worker_ids.includes(worker.worker_id)) {
        workers.push(worker);
      }
    });
    return workers;
  }

  render() {
    return (
      <section>
        <header>Processes</header>
        <p>{this.state.workerPool.length} Worker process(es). {this.state.busyWorker.length} active worker(s) out of {this.workerCount}.</p>
        {
          this.state.workerPool.map((pool) => {
            let busyWorker = this.getBusyPoolWorker(pool);
            return (
              <div key={pool.worker_pool_id} className={cx(styles.panel, styles.panelDefault)}>
                <div className={styles.tableResponsive}>
                  <table className={styles.table}>
                    <tbody>
                      <tr>
                        <td>{pool.host}: {pool.pid}</td>
                        <td>Started <UnixTime ts={pool.started_at}/></td>
                        <td>Last Heartbeat <UnixTime ts={pool.heartbeat_at}/></td>
                        <td>Concurrency {pool.concurrency}</td>
                      </tr>
                      <tr>
                        <td colSpan="4">Servicing <ShortList item={pool.job_names} />.</td>
                      </tr>
                      <tr>
                        <td colSpan="4">{busyWorker.length} active worker(s) and {pool.worker_ids.length - busyWorker.length} idle.</td>
                      </tr>
                      <tr>
                        <td colSpan="4">
                          <div className={cx(styles.panel, styles.panelDefault)}>
                            <BusyWorkers worker={busyWorker} />
                          </div>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            );
          })
        }
      </section>
    );
  }
}
