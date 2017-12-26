# Customize the look & feel of Harbor

The primary look & feel of Harbor supports to be customized with several simple steps. All the relevant customization configurations are saved in the `setting.json` file under `$HARBOR_DIR/src/ui_ng/src` folder with `json` format and will be loaded when the application is started to run. 

## Configure
Open the `setting.json` file, you'll see the default content as shown below:

```
{
  "headerBgColor": "#004a70",
  "headerLogo": "",
  "loginBgImg": "",
  "product": {
    "title": "VMware Harbor",
    "company": "vmware",
    "name": "Harbor",
    "introductions": {
      "zh-cn": "",
      "es-es": "",
      "en-us": ""
    }
  }
}
```

Change the configuration value if you want to override the default style to your own. Here are references:

* **headerBgColor**: Background color of the page header, both hex and rgb formats support. e.g: `#004a70` and `rgb(210,110,235)`.
* **headerLogo**: Relevant path of the logo image in the header. The image file should be put in the `images` folder.
* **loginBgImg**: Background image displayed in the login page. The image file should be put in the `images` folder. Suggest the size of this image should be bigger than 800px*600px.
* **Product**: Contain metadata / description of the product.
  - **title**: The full product title displayed in the login page.
  - **company**: Name of the company publishing the product.
  - **name**: Name of the product.
  - **introductions**: The introduction texts about the product with different languages, which are displayed in the `About` dialog.

## Build

Once the `setting.json` configurations changed, re-[build](#configure) your product to apply the new changes.