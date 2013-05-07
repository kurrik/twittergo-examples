module.exports = function(grunt) {
  grunt.initConfig({
    pkg: grunt.file.readJSON('package.json'),

    bgShell: {
      serve: {
        cmd: 'pkill -f dev_appserver; ' +
             '~/src/google_appengine_go/dev_appserver.py ' +
             '--port=9996 --address=0.0.0.0 src',
        bg: false,
      },
      deploy: {
        cmd: '~/src/google_appengine_go/appcfg.py --oauth2 update src',
        bg: false,
      },
    },
  });
  grunt.loadNpmTasks('grunt-bg-shell');
  grunt.registerTask('develop', ['bgShell:serve']);
  grunt.registerTask('deploy', ['bgShell:deploy']);
};
