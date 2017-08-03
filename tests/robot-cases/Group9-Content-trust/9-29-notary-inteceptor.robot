#Test Case - Notary Inteceptor
#    ${rc}  ${output}=  Run And Return Rc And Output  ./tests/robot-cases/Group9-Content-trust/notary-pull-image-inteceptor.sh
#    Log To Console  ${output}
#    Should Be Equal As Integers  ${rc}  0
#
#    Down Harbor  with_notary=true
#    ${rc}  ${output}=  Run And Return Rc And Output  echo "PROJECT_CONTENT_TRUST=1\n" >> ./make/common/config/ui/env
#    Log To Console  ${output}
#    Should Be Equal As Integers  ${rc}  0
#    ${rc}  ${output}=  Run And Return Rc And Output  cat ./make/common/config/ui/env
#
#    Log To Console  ${output}
#	Up Harbor  with_notary=true
#    ${rc}  ${output}=  Run And Return Rc And Output  ./tests/robot-cases/Group9-Content-trust/notary-pull-image-inteceptor.sh
#    Log To Console  ${output}
#
#	Down Harbor  with_notary=true
#	${rc}  ${output}=  Run And Return Rc And Output  sed "s/^PROJECT_CONTENT_TRUST=1.*/PROJECT_CONTENT_TRUST=0/g" -i ./make/common/config/ui/env
#   Log To Console  ${output}
#   Should Be Equal As Integers  ${rc}  0
#    ${rc}  ${output}=  Run And Return Rc And Output  cat ./make/common/config/ui/env
#
#	Up Harbor  with_notary=true
#    ${rc}  ${output}=  Run And Return Rc And Output  ./tests/robot-cases/Group9-Content-trust/notary-pull-image-inteceptor.sh
#    Log To Console  ${output}
#    Should Be Equal As Integers  ${rc}  0