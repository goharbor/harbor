const headerPlugin = require('eslint-plugin-header');

// eslint-plugin-header 3.1.1 does not declare a rule schema, which ESLint 9
// requires. Patching it here avoids the need for patch-package.
headerPlugin.rules.header.meta.schema = { type: 'array', items: {}, minItems: 0 };

module.exports = {
  root: true,
  ignorePatterns: [
    "projects/**/*",
    "**/*.js"
  ],
  overrides: [
    {
      files: ["*.ts"],
      parserOptions: {
        project: [
          "server/tsconfig.json",
          "tsconfig.json",
          "cypress/tsconfig.json"
        ],
        createDefaultProgram: true
      },
      extends: [
        "plugin:@angular-eslint/recommended",
        "plugin:@angular-eslint/template/process-inline-templates",
        "plugin:prettier/recommended"
      ],
      rules: {
        "no-console": ["error", { allow: ["warn", "error"] }],
        "@angular-eslint/prefer-standalone": "off",
        "@angular-eslint/no-output-native": "off",
        "@angular-eslint/prefer-inject": "off"
      }
    },
    {
      files: ["*.html"],
      extends: [
        "plugin:@angular-eslint/template/recommended"
      ],
      rules: {
        "@angular-eslint/template/prefer-control-flow": "off",
        "@angular-eslint/template/prefer-self-closing-tags": "off"
      }
    },
    {
      files: ["*.html"],
      excludedFiles: ["*inline-template-*.component.html"],
      extends: ["plugin:prettier/recommended"],
      rules: {
        "prettier/prettier": ["error", { parser: "angular" }]
      }
    },
    {
      files: ["src/**/*.ts"],
      plugins: ["header"],
      rules: {
        "header/header": [2, "./copyright.tmpl.js"]
      }
    }
  ]
};
