module.exports = parameters => {
  for (const param of parameters) {
    if (param.name === 'X-Request-Id' && param.in === 'header') {
      return
    }
  }

  return [
    {
      message: 'X-Request-Id must be in "parameters".',
    },
  ];
};
