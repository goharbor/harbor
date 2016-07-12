## Developer's Guide for Internationalization (i18n)

*NOTE: All the files you created should use UTF-8 encoding.*

### Steps to localize the UI in your language

1. Copy the file `static/resources/js/services/i18n/locale_messages_en-US.js` to a new file in the same directory named `locale_messages_<language>_<locale>.js` .

    The file contains a JSON object named `locale_messages`, which consists of key-value pairs of UI strings:
    ```
    var local_messages = {
       'sign_in': 'Sign In',
       'sign_up': 'Sign Up',
          ...
     };
    ```  
    In the file `locale_messages_<language>_<locale>.js`, translate all the values into your language. Do not change any keys.

2. After creating your locale file, you should include it from the HTML page header template.

    In the file `views/sections/header-include.htm`, look for a `if` statement which switch langauges based on the current language (`.Lang`) value. Add in a `else if` statement for your language:
    ```
    {{ if eq .Lang "zh-CN" }}
	   <script src="/static/resources/js/services/i18n/locale_messages_zh-CN.js"></script>
	{{ else if eq .Lang "en-US"}}
	   <script src="/static/resources/js/services/i18n/locale_messages_en-US.js"></script>
   ** {{ else if eq .Lang "<language>-<locale>"}}
       <script src="/static/resources/js/services/i18n/locale_messages_<language>-<locale>.js"></script>**
	{{ end }}
    ```
3. Add the new language to the `I18nService` module.

    In the file `static/resources/js/services/i18n/services.i18n.js`, append a new key-value item to the `supportLanguages` object. This value will be displayed in the language dropdown list in the UI.
    ```
     var supportLanguages = {
      'en-US': 'English',
      'zh-CN': '中文',
      '<language>-<locale>': '<language_name>'
     };
    ```
    **NOTE: Don't miss the comma before the new key-value item you've added.**


4. In the directory `static/i18n/`, copy the file `locale_en-US.ini` to a new file named  `locale_<language>-<locale>.ini`. In this file, translate all the values on the right hand side into your language. Do not change any keys.

5. Add the new language to the `app.conf` file.
    
    In the file `Deploy/config/ui/app.conf`, append a new item to the configuration section.
    ```
		 [lang]
		 types = en-US|zh-CN|<language>-<locale>
		 names = en-US|zh-CN|<language>-<locale>
    ```

6. Next, change to `Deploy/` directory, rebuild and restart the Harbor by the below command: 
    ```
        docker-compose down
        docker-compose up --build -d
    ```
    