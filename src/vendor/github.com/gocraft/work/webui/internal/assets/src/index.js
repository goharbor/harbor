import React from 'react';
import PropTypes from 'prop-types';
import { render } from 'react-dom';
import Processes from './Processes';
import DeadJobs from './DeadJobs';
import Queues from './Queues';
import RetryJobs from './RetryJobs';
import ScheduledJobs from './ScheduledJobs';
import { Router, Route, Link, IndexRedirect, hashHistory } from 'react-router';
import styles from './bootstrap.min.css';
import cx from './cx';

class App extends React.Component {
  static propTypes = {
    children: PropTypes.element.isRequired,
  }

  render() {
    return (
      <div className={styles.container} style={{marginTop: 30, marginBottom: 60}}>
        <header><h1>gocraft/work</h1></header>
        <hr />
        <div className={styles.row}>
          <main className={styles.colMd10}>
            {this.props.children}
          </main>
          <aside className={styles.colMd2}>
            <nav>
              <ul className={cx(styles.nav, styles.navPills, styles.navStacked)}>
                <li><Link to="/processes">Processes</Link></li>
                <li><Link to="/queues">Queues</Link></li>
                <li><Link to="/retry_jobs">Retry Jobs</Link></li>
                <li><Link to="/scheduled_jobs">Scheduled Jobs</Link></li>
                <li><Link to="/dead_jobs">Dead Jobs</Link></li>
              </ul>
            </nav>
          </aside>
        </div>
      </div>
    );
  }
}

// react-router's route cannot be used to specify props to children component.
// See https://github.com/reactjs/react-router/issues/1857.
render(
  <Router history={hashHistory}>
    <Route path="/" component={App}>
      <Route path="/processes" component={ () => <Processes busyWorkerURL="/busy_workers" workerPoolURL="/worker_pools" /> } />
      <Route path="/queues" component={ () => <Queues url="/queues" /> } />
      <Route path="/retry_jobs" component={ () => <RetryJobs url="/retry_jobs" /> } />
      <Route path="/scheduled_jobs" component={ () => <ScheduledJobs url="/scheduled_jobs" /> } />
      <Route path="/dead_jobs" component={ () =>
        <DeadJobs
          fetchURL="/dead_jobs"
          retryURL="/retry_dead_job"
          retryAllURL="/retry_all_dead_jobs"
          deleteURL="/delete_dead_job"
          deleteAllURL="/delete_all_dead_jobs"
        />
      } />
      <IndexRedirect from="" to="/processes" />
    </Route>
  </Router>,
  document.getElementById('app')
);
