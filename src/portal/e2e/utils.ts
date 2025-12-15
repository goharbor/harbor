const { exec } = require('child_process');
const util = require('util');
const execAsync = util.promisify(exec);

export async function runCommand(cmd) {
  try {
    const { stdout, stderr } = await execAsync(cmd);
    return stdout + stderr;
  } catch (error) {
    return error.stdout + error.stderr;
  }
}

export async function pullImage({ ip, user, pwd, project, image, tag = null, isRobot = false }) {
  console.log(`\nRunning docker pull ${image}...`);
  const imageWithTag = tag === null ? image : `${image}:${tag}`;
  const loginCmd = isRobot
    ? `docker login -u robot\\$${project}+${user} -p ${pwd} ${ip}`
    : `docker login -u ${user} -p ${pwd} ${ip}`;
  await runCommand(loginCmd);

  const pullCmd = `docker pull ${ip}/${project}/${imageWithTag}`;
  const output = await runCommand(pullCmd);
  console.log(output);

  if (!output.includes('Digest:')) throw new Error('Output missing Digest');
  if (!output.includes('Status:')) throw new Error('Output missing Status');
  if (output.includes('No such image:')) throw new Error('Image not found');
}

export async function pushImage({
  ip, user, pwd, project, imageWithOrWithoutTag,
  needPullFirst = true, sha256 = null, isRobot = false,
  localRegistry, localRegistryNamespace
}) {
  const d = Date.now();
  const imageInUse = sha256 === null
    ? imageWithOrWithoutTag
    : `${imageWithOrWithoutTag}@sha256:${sha256}`;
  const imageInUseWithTag = sha256 === null
    ? imageWithOrWithoutTag
    : `${imageWithOrWithoutTag}:${sha256}`;

  await new Promise(r => setTimeout(r, 3000));
  console.log(`\nRunning docker push ${imageWithOrWithoutTag}...`);

  let imageToTag = imageWithOrWithoutTag;
  if (needPullFirst) {
    await runCommand(`docker pull ${localRegistry}/${localRegistryNamespace}/${imageInUse}`);
    imageToTag = imageInUse;
  }

  const loginCmd = isRobot
    ? `docker login -u robot\\$${project}+${user} -p ${pwd} ${ip}`
    : `docker login -u ${user} -p ${pwd} ${ip}`;
  await runCommand(loginCmd);

  if (needPullFirst) {
    await runCommand(`docker tag ${localRegistry}/${localRegistryNamespace}/${imageToTag} ${ip}/${project}/${imageInUseWithTag}`);
  } else {
    await runCommand(`docker tag ${imageToTag} ${ip}/${project}/${imageInUseWithTag}`);
  }

  await runCommand(`docker push ${ip}/${project}/${imageInUseWithTag}`);
  await runCommand(`docker logout ${ip}`);
  await new Promise(r => setTimeout(r, 1000));
}

export async function pushImageWithTag({
  ip, user, pwd, project, image, tag, tag1 = 'latest',
  localRegistry, localRegistryNamespace
}) {
  console.log(`\nRunning docker push ${image}...`);
  await runCommand(`docker pull ${localRegistry}/${localRegistryNamespace}/${image}:${tag1}`);
  await runCommand(`docker login -u ${user} -p ${pwd} ${ip}`);
  await runCommand(`docker tag ${localRegistry}/${localRegistryNamespace}/${image}:${tag1} ${ip}/${project}/${image}:${tag}`);
  await runCommand(`docker push ${ip}/${project}/${image}:${tag}`);
  await runCommand(`docker logout ${ip}`);
}