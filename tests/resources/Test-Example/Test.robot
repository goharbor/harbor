
*** Settings ***
Documentation  Harbor BATs
Resource  ../Util.robot

*** Variables ***
${HARBOR_URL}  https://${ip}
${HARBOR_ADMIN}  admin
${HARBOR_PASSWORD}  Harbor12345

*** Test Cases ***
# For Windows
Test Case - Example For Windows
    Open Browser    https://localhost:4200    Chrome
    Retry Element Click  xpath=//clr-dropdown/button
    Retry Element Click  xpath=//clr-dropdown/clr-dropdown-menu/a[contains(., 'English')]
    # your case starts =====================================

    # your case ends ======================================
    Close Browser

# For Linux
Test Case - Example For Linux
    init chrome driver
    # your case starts =====================================

    # your case ends ======================================
    Close Browser