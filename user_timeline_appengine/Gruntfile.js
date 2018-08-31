module.exports = function(grunt) {
  grunt.initConfig({
    pkg: grunt.file.readJSON('package.json'),

    bgShell: {
      serve: {
        cmd: 'pkill -f dev_appserver; ' +
             '~/src/go_appengine/dev_appserver.py ' +
             '--port=9996 src',
        bg: false,
      },
      deploy: {
        cmd: '~/src/go_appengine/appcfg.py --oauth2 update src',
        bg: false,
      },
    },
  });
  grunt.loadNpmTasks('grunt-bg-shell');
  grunt.registerTask('develop', ['bgShell:serve']);
  grunt.registerTask('deploy', ['bgShell:deploy']);
};
