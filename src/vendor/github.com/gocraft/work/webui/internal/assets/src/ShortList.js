import React from 'react';
import styles from './ShortList.css';

export default class ShortList extends React.Component {
  static propTypes = {
    item: React.PropTypes.arrayOf(React.PropTypes.string).isRequired,
  }

  render() {
    return (
      <ul className={styles.ul}>
        {
          this.props.item.map((item, i) => {
            if (i < 3) {
              return (<li key={i} className={styles.li}>{item}</li>);
            } else if (i == 3) {
              return (<li key={i} className={styles.li}>{this.props.item.length - 3} more</li>);
            }
          })
        }
      </ul>
    );
  }
}
