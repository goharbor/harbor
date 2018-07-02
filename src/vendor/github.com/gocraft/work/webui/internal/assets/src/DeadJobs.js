import React from 'react';
import PropTypes from 'prop-types';
import PageList from './PageList';
import UnixTime from './UnixTime';
import styles from './bootstrap.min.css';
import cx from './cx';

export default class DeadJobs extends React.Component {
  static propTypes = {
    fetchURL: PropTypes.string,
    deleteURL: PropTypes.string,
    deleteAllURL: PropTypes.string,
    retryURL: PropTypes.string,
    retryAllURL: PropTypes.string,
  }

  state = {
    selected: [],
    page: 1,
    count: 0,
    jobs: []
  }

  fetch() {
    if (!this.props.fetchURL) {
      return;
    }
    fetch(`${this.props.fetchURL}?page=${this.state.page}`).
      then((resp) => resp.json()).
      then((data) => {
        this.setState({
          selected: [],
          count: data.count,
          jobs: data.jobs
        });
      });
  }

  componentWillMount() {
    this.fetch();
  }

  updatePage(page) {
    this.setState({page: page}, this.fetch);
  }

  checked(job) {
    return this.state.selected.includes(job);
  }

  check(job) {
    var index = this.state.selected.indexOf(job);
    if (index >= 0) {
      this.state.selected.splice(index, 1);
    } else {
      this.state.selected.push(job);
    }
    this.setState({
      selected: this.state.selected
    });
  }

  checkAll() {
    if (this.state.selected.length > 0) {
      this.setState({selected: []});
    } else {
      this.state.jobs.map((job) => {
        this.state.selected.push(job);
      });
      this.setState({
        selected: this.state.selected
      });
    }
  }

  deleteAll() {
    if (!this.props.deleteAllURL) {
      return;
    }
    fetch(this.props.deleteAllURL, {method: 'post'}).then(() => {
      this.updatePage(1);
    });
  }

  deleteSelected() {
    let p = [];
    this.state.selected.map((job) => {
      if (!this.props.deleteURL) {
        return;
      }
      p.push(fetch(`${this.props.deleteURL}/${job.died_at}/${job.id}`, {method: 'post'}));
    });

    Promise.all(p).then(() => {
      this.fetch();
    });
  }

  retryAll() {
    if (!this.props.retryAllURL) {
      return;
    }
    fetch(this.props.retryAllURL, {method: 'post'}).then(() => {
      this.updatePage(1);
    });
  }

  retrySelected() {
    let p = [];
    this.state.selected.map((job) => {
      if (!this.props.retryURL) {
        return;
      }
      p.push(fetch(`${this.props.retryURL}/${job.died_at}/${job.id}`, {method: 'post'}));
    });

    Promise.all(p).then(() => {
      this.fetch();
    });
  }

  render() {
    return (
      <div>
        <div className={cx(styles.panel, styles.panelDefault)}>
          <div className={styles.panelHeading}>Dead Jobs</div>
          <div className={styles.panelBody}>
            <p>{this.state.count} job(s) are dead.</p>
            <PageList page={this.state.page} totalCount={this.state.count} perPage={20} jumpTo={(page) => () => this.updatePage(page)}/>
          </div>
          <div className={styles.tableResponsive}>
            <table className={styles.table}>
              <tbody>
                <tr>
                  <th><input type="checkbox" checked={this.state.selected.length > 0} onChange={() => this.checkAll()}/></th>
                  <th>Name</th>
                  <th>Arguments</th>
                  <th>Error</th>
                  <th>Died At</th>
                </tr>
                {
                  this.state.jobs.map((job) => {
                    return (
                      <tr key={job.id}>
                        <td><input type="checkbox" checked={this.checked(job)} onChange={() => this.check(job)}/></td>
                        <td>{job.name}</td>
                        <td>{JSON.stringify(job.args)}</td>
                        <td>{job.err}</td>
                        <td><UnixTime ts={job.t} /></td>
                      </tr>
                    );
                  })
                }
              </tbody>
            </table>
          </div>
        </div>
        <div className={styles.btnGroup} role="group">
          <button type="button" className={cx(styles.btn, styles.btnDefault)} onClick={() => this.deleteSelected()}>Delete Selected Jobs</button>
          <button type="button" className={cx(styles.btn, styles.btnDefault)} onClick={() => this.retrySelected()}>Retry Selected Jobs</button>
          <button type="button" className={cx(styles.btn, styles.btnDefault)} onClick={() => this.deleteAll()}>Delete All Jobs</button>
          <button type="button" className={cx(styles.btn, styles.btnDefault)} onClick={() => this.retryAll()}>Retry All Jobs</button>
        </div>
      </div>
    );
  }
}
