/*global module:false*/
module.exports = function(grunt) {

  'use strict';
  // Project configuration.
  grunt.initConfig({
    // Task configuration.
    jshint: {
      options: {
        browser: true,
        curly: true,
        freeze: true,
        bitwise: true,
        eqeqeq: true,
        strict: true,
        immed: true,
        latedef: false,
        newcap: false,
        smarttabs: true,
        noarg: true,
        devel: true,
        sub: true,
        undef: true,
        unused: false,
        boss: true,
        eqnull: true,
        globals: {
          jQuery: true,
          angular: true,
          $: true,
        }
      },
      gruntfile: {
        src: 'Gruntfile.js'
      },
      scripts: {
        src: ['resources/**/**/*.js']
      }
    },
    watch: {
      gruntfile: {
        files: '<%= jshint.gruntfile.src %>',
        tasks: ['jshint:gruntfile']
      },
      scripts: {
        files: '<%= jshint.scripts.src %>',
        tasks: ['jshint:scripts']
      }
    }
  });

  // These plugins provide necessary tasks.
  grunt.loadNpmTasks('grunt-contrib-jshint');
  grunt.loadNpmTasks('grunt-contrib-watch');

  // Default task.
  grunt.registerTask('default', ['jshint']);

};
