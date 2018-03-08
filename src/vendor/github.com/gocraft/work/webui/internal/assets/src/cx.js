
// react only supports className as a string.
// See https://github.com/facebook/react/pull/1198.
export default (...classNames) => {
  return classNames.join(' ');
};
