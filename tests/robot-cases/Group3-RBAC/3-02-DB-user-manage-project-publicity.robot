Test Case Manage project publicity
    #Start Docker Daemon Locally
    Init Chrome Driver
    ${d}=    Get Current Date  result_format=%m%s
    ${rc}  ${ip}=    run and return rc and output  ip a s eth0|grep "inet "|awk '{print $2}'|awk -F "/" '{print $1}'
    Log to console  ${ip}

    Create An New User  url=${HARBOR_URL}  username=usera${d}  email=usera${d}@vmware.com  realname=usera${d}  newPassword=Test1@34  comment=harbor
    Logout Harbor
    Create An New User  url=${HARBOR_URL}  username=userb${d}  email=userb${d}@vmware.com  realname=userb${d}  newPassword=Test1@34  comment=harbor
    Logout Harbor

    Sign In Harbor  ${HARBOR_URL}  usera${d}  Test1@34
    Create An New Public Project  project${d}

    Push image  10.112.122.188  usera${d}  Test1@34  project${d}  hello-world:latest
    Pull image  10.112.122.188  userb${d}  Test1@34  project${d}  hello-world:latest

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  userb${d}  Test1@34
    Project Should Display  project${d}
    Search Private Projects
    Project Should Not Display  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  usera${d}  Test1@34
    Make Project Private  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  userb${d}  Test1@34
    Project Should Not Display  project${d}
    Cannot Pull image  10.112.122.188  usera${d}  Test1@34  project${d}  hello-world:latest

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  usera${d}  Test1@34
    Make Project Public  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  userb${d}  Test1@34
    Project Should Display  project${d}
