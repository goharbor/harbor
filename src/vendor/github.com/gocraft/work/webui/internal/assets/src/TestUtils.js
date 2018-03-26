
export let walkOutput = (root, walkFunc) => {
  if (root == undefined) {
    return;
  }
  if (Array.isArray(root)) {
    root.map((el) => {
      walkOutput(el, walkFunc);
    });
    return;
  }

  if (root.type) {
    walkFunc(root);
    walkOutput(root.props.children, walkFunc);
  }
};

export let findAllByTag = (root, tag) => {
  let found = [];
  walkOutput(root, (el) => {
    if (typeof el.type === 'function' && el.type.name == tag) {
      found.push(el);
    } else if (el.type === tag) {
      found.push(el);
    }
  });
  return found;
};
