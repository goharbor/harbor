## Guide to Harbor E2E testing

### Prerequisites

In order to run e2e testing, you need to install the following tools:

1. Elastic search  -- required ES_ENDPOINT
2. Dex -- OIDC testing environment -- required by OIDC login testcase, if not used, just ignore it.
3. LDAP server -- required by LDAP login testcase. if not used, just ignore it.
4. Webhook Server -- required by webhook testcase. WEBHOOK_ENDPOINT
5. Dragonfly  -- Required by p2p preheat testcase, DISTRIBUTION_ENDPOINT,  DRAGONFLY_AUTH_TOKEN
6. Fake Scanner -- required by scanner testcase, SCANNER_ENDPOINT
7. Install Harbor, need to installed with https enabled. the root ca should be the root ca in harbor repository: https://github.com/goharbor/harbor/blob/main/tests/harbor_ca.crt




#### 1. Git clone the following repositories:

```
    git clone https://github.com/goharbor/harbor
```

#### 2. Update the tests/e2e_setup/robotvars.py in e2e_setup, update the environment variable with the prerequest env settings. change the ip to the IP address of the harbor instance

```
cd tests/e2e_setup
cp robotvars.sample.py robotvars.py
# update the environment variable in robotvars.py

```

#### 3. Start e2e container

```
./e2e_container.sh
```

Check if harbor_ca.crt exists in /ca directory. if not copy it
```
cp /ca/ca.crt /ca/harbor_ca.crt
```

#### 4. Run setup, in the previous container console, run the following command.
```
robot -V /drone/tests/e2e_setup/robotvars.py /drone/tests/robot-cases/Group1-Nightly/Setup_Nightly.robot
```

#### 5. Run robot test

After setup you can select to run single test or full test

##### 5.1 Run single test
```
robot --include sbom_manual_gen -V /drone/tests/e2e_setup/robotvars.py /drone/tests/robot-cases/Group1-Nightly/Trivy.robot
```

##### 5.2 Run full test
```
robot -V /drone/tests/e2e_setup/robotvars.py  /drone/tests/robot-cases/Group1-Nightly/Setup_Nightly.robot /drone/tests/robot-cases/Group1-Nightly/Common_GC.robot /drone/tests/robot-cases/Group1-Nightly/Webhook.robot /drone/tests/robot-cases/Group1-Nightly/Routing.robot /drone/tests/robot-cases/Group1-Nightly/P2P_Preheat.robot /drone/tests/robot-cases/Group1-Nightly/Trivy.robot /drone/tests/robot-cases/Group1-Nightly/DB.robot /drone/tests/robot-cases/Group1-Nightly/Common.robot /drone/tests/robot-cases/Group1-Nightly/Teardown.robot
```

#### 6. Check report in the harbor source code directory.

After the test complete, check the test report in the harbor source code directory.

