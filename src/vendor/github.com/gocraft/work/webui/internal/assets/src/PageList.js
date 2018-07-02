import React from 'react';
import PropTypes from 'prop-types';
import styles from './bootstrap.min.css';

export default class PageList extends React.Component {
  static propTypes = {
    page: PropTypes.number.isRequired,
    perPage: PropTypes.number.isRequired,
    totalCount: PropTypes.number.isRequired,
    jumpTo: PropTypes.func.isRequired,
  }

  get totalPage() {
    return Math.ceil(this.props.totalCount / this.props.perPage);
  }

  shouldShow(i) {
    if (i == 1 || i == this.totalPage) {
      return true;
    }
    return Math.abs(this.props.page - i) <= 1;
  }

  render() {
    if (this.totalPage == 0) {
      return null;
    }
    let pages = [];
    for (let i = 1; i <= this.totalPage; i++) {
      if (i == this.props.page) {
        pages.push(<li key={i} className={styles.active}><span>{i}</span></li>);
      } else if (this.shouldShow(i)) {
        pages.push(<li key={i}><a onClick={this.props.jumpTo(i)}>{i}</a></li>);
      } else if (this.shouldShow(i-1)) {
        pages.push(<li key={i} className={styles.disabled}><span>..</span></li>);
      }
    }
    return (
      <ul className={styles.pagination}>{pages}</ul>
    );
  }
}
