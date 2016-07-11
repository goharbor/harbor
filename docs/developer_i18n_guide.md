# Developer's guide for internationalization

*NOTE: All the files you created should be in UTF-8 encoding.*

>**For the front-end i18n works,**

1. You need to create a JavaScript i18n source file under the diretory `/static/resources/js/services/i18n`, it should be named as `locale_messages_<language>_<locale>.js` with your localized translations. 

    This file contains a JSON object named `locale_messages`, and the sample pattern should be like below,
    ```
    var local_messages = {
    	'sign_in': 'Sign In'
     };
    ```  
    **NOTE: Please refer the keys which already exist in other i18n files to make your corresponding translations as your locale.**

2. After creating this locale file, you should include it into the HTML page header template.

    In the file, `/views/sections/header-include.htm`. This template would be rendered by the back-end controller, with checking the current language (`.Lang`) value from the request scope, at each time there is only one script tag would be rendered at front-end page.
    ```
    {{ if eq .Lang "zh-CN" }}
	   <script src="/static/resources/js/services/i18n/locale_messages_zh-CN.js"></script>
	{{ else if eq .Lang "en-US"}}
	   <script src="/static/resources/js/services/i18n/locale_messages_en-US.js"></script>
    {{ else if eq .Lang "<language>-<locale>"}}
       <script src="/static/resources/js/services/i18n/locale_messages_<language>-<locale>.js"></script>
	{{ end }}
    ```
3. Add the new coming language to the `I18nService` module.

    In the file, `/static/resources/js/services/i18n/services.i18n.js`, append new key-value item to the `supportLanguages` object.
    ```
     var supportLanguages = {
      'en-US': 'English',
      'zh-CN': '中文',
      '<language>-<locale>': '<language_name>'
     };
    ```
    **NOTE: Don't miss to add a comma ahead of the new key-value item you've added.**

>**For the back-end i18n works,**

1. Create a file under the directory `/static/i18n/`, named as `locale_<language>-<locale>.ini`.

    **NOTE: Please refer the keys which already exist in other i18n files to make your corresponding translations as your locale.**

2. Add the new comming language to the `app.conf` file.
    
    In the file, `/Deploy/config/ui/app.conf`, append new item to the configuration section.
    ```
		 [lang]
		 types = en-US|zh-CN|<language>-<locale>
		 names = en-US|zh-CN|<language>-<locale>
    ```
>**Rebuild and start the Harbor project by using 'docker-compose' command.**