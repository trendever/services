'use strict';

var gulp = require('gulp');
var eslint = require('gulp-eslint');
var plugins = require('gulp-load-plugins')();

var fs = require('fs');
var path = require('path');
var es = require('event-stream');
var rename = require('gulp-rename');

var moduleName = (function () {
      var args = process.argv;
      var length = args.length;
      var i = 0;
      var name;
      var subName;

      while (i++ < length) {
        if ((/^--+(\w+)/i).test(args[i])){
          name = args[i].split('--')[1];
          subName = args[i].split('--')[2];
          break;
        }
      }
      return {
        'name': name,
        'subName': subName
      };
    })();

// Admin Module
// Command: gulp [task]
// Admin is default task
// Watch Admin module: gulp
// -----------------------------------------------------------------------------

function adminTasks() {
  var pathto = function (file) {
        return ('../admin/views/assets/' + file);
      };
  var scripts = {
        src: pathto('javascripts/app/*.js'),
        dest: pathto('javascripts'),
        qor: pathto('javascripts/qor/*.js'),
        qorInit: pathto('javascripts/qor/qor-init.js'),
        all: [
          'gulpfile.js',
          pathto('javascripts/qor/*.js')
        ]
      };
  var styles = {
        src: pathto('stylesheets/scss/{app,qor}.scss'),
        dest: pathto('stylesheets'),
        vendors: pathto('stylesheets/vendors'),
        main: pathto('stylesheets/{qor,app}.css'),
        scss: pathto('stylesheets/scss/**/*.scss')
      };

  gulp.task('qor', function () {
    return gulp.src([scripts.qorInit,scripts.qor])
    .pipe(plugins.concat('qor.js'))
    .pipe(plugins.uglify())
    .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('js', ['qor'], function () {
    return gulp.src(scripts.src)
    .pipe(eslint({configFile: '.eslintrc'}))
    .pipe(eslint.format())
    .pipe(plugins.concat('app.js'))
    .pipe(plugins.uglify())
    .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('qor+', function () {
    return gulp.src([scripts.qorInit,scripts.qor])
    .pipe(eslint({configFile: '.eslintrc'}))
    .pipe(eslint.format())
    .pipe(plugins.concat('qor.js'))
    .pipe(plugins.uglify())
    .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('js+', function () {
    return gulp.src(scripts.src)
    .pipe(plugins.concat('app.js'))
    .pipe(plugins.uglify())
    .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('sass', function () {
    return gulp.src(styles.src)
    .pipe(plugins.sass())
    .pipe(gulp.dest(styles.dest));
  });

  gulp.task('css', ['sass'], function () {
    return gulp.src(styles.main)
    .pipe(plugins.autoprefixer())
    .pipe(plugins.csscomb())
    .pipe(plugins.minifyCss())
    .pipe(gulp.dest(styles.dest));
  });

  gulp.task('watch', function () {
    gulp.watch(scripts.qor, ['qor+']);
    gulp.watch(scripts.src, ['js+']);
    gulp.watch(styles.scss, ['css']);
  });

  gulp.task('release', ['js', 'css']);

  gulp.task('default', ['watch']);
}


// -----------------------------------------------------------------------------
// Other Modules
// Command: gulp [task] --moduleName--subModuleName
//
//  example:
// Watch Worker module: gulp --worker
//
// if module's assets just as normal path:
// moduleName/views/themes/moduleName/assets/javascripts(stylesheets)
// just use gulp --worker
//
// if module's assets in enterprise as normal path:
// moduleName/views/themes/moduleName/assets/javascripts(stylesheets)
// just use gulp --microsite--enterprise
//
// if module's assets path as Admin module:
// moduleName/views/assets/javascripts(stylesheets)
// you need set subModuleName as admin
// gulp --worker--admin
//
// if you need run task for subModule in modules
// example: worker module inline_edit subModule:
// gulp --worker--inline_edit
//
// -----------------------------------------------------------------------------

function moduleTasks(moduleNames) {
  var moduleName = moduleNames.name;
  var subModuleName = moduleNames.subName;

  var pathto = function (file) {
    if(moduleName && subModuleName) {
      if(subModuleName == 'admin') {
        return '../' + moduleName + '/views/assets/' + file;
      } else if (subModuleName == 'enterprise'){
        return '../../../enterprise.getqor.com/' + moduleName + '/views/themes/' + moduleName + '/assets/' + file;
      } else {
        return '../' + moduleName + '/' + subModuleName + '/views/themes/' + moduleName + '/assets/' + file;
      }

    }
    return '../' + moduleName + '/views/themes/' + moduleName + '/assets/' + file;
  };

  var scripts = {
        src: pathto('javascripts/'),
        watch: pathto('javascripts/**/*.js')
      };
  var styles = {
        src: pathto('stylesheets/'),
        watch: pathto('stylesheets/**/*.scss')
      };

  function getFolders(dir){
    return fs.readdirSync(dir).filter(function(file){
      return fs.statSync(path.join(dir, file)).isDirectory();
    })
  }


  gulp.task('js', function () {
    var scriptPath = scripts.src;
    var folders = getFolders(scriptPath);
    var task = folders.map(function(folder){

      return gulp.src(path.join(scriptPath, folder, '/*.js'))
        .pipe(eslint({configFile: '.eslintrc'}))
        .pipe(eslint.format())
        .pipe(plugins.concat(folder + '.js'))
        .pipe(plugins.uglify())
        .pipe(gulp.dest(scriptPath));
    });

    return es.concat.apply(null, task);

  });

  gulp.task('css', function () {

    var stylePath = styles.src;
    var folders = getFolders(stylePath);
    var task = folders.map(function(folder){

      return gulp.src(path.join(stylePath, folder, '/*.scss'))

        .pipe(plugins.sass({outputStyle: 'compressed'}))
        .pipe(plugins.minifyCss())
        .pipe(rename(folder + '.css'))
        .pipe(gulp.dest(stylePath))
    });

    return es.concat.apply(null, task);

  });

  gulp.task('watch', function () {
    gulp.watch(scripts.watch, ['js']);
    gulp.watch(styles.watch, ['css']);
  });

  gulp.task('default', ['watch']);
  gulp.task('release', ['js', 'css']);
}


// Init
// -----------------------------------------------------------------------------

if (moduleName.name) {
  var taskPath = moduleName.name + '/views/themes/' + moduleName.name + '/assets/';
  var runModuleName = 'Running "' + moduleName.name + '" module task in "' + taskPath + '"...';

  if (moduleName.subName){
    if (moduleName.subName == 'admin'){
      taskPath = moduleName.name + '/views/assets/';
      runModuleName = 'Running "' + moduleName.name + '" module task in "' + taskPath + '"...';
    } else if (moduleName.subName == 'enterprise'){
      taskPath = '../../../enterprise.getqor.com/' + moduleName.name + '/views/themes/' + moduleName.name + '/assets/';
      runModuleName = 'Running "' + moduleName.name + '" module task in "' + taskPath + '"...';
    }else {
      taskPath = moduleName.name + '/' + moduleName.subName + '/views/themes/' + moduleName.name + '/assets/';
      runModuleName = 'Running "' + moduleName.name + ' > ' + moduleName.subName + '" module task in "' + taskPath + '"...';
    }
  }
  console.log(runModuleName);
  moduleTasks(moduleName);
} else {
  console.log('Running "admin" module task in "admin/views/assets/"...');
  adminTasks();
}

// Task for compress js and css vendor assets
gulp.task('compressJavaScriptVendor', function () {
  return gulp.src(['!../admin/views/assets/javascripts/vendors/jquery.min.js','../admin/views/assets/javascripts/vendors/*.js'])
  .pipe(plugins.uglify())
  .pipe(plugins.concat('vendors.js'))
  .pipe(gulp.dest('../admin/views/assets/javascripts'));
});

gulp.task('compressCSSVendor', function () {
  return gulp.src('../admin/views/assets/stylesheets/vendors/*.css')
  .pipe(plugins.concat('vendors.css'))
  .pipe(gulp.dest('../admin/views/assets/stylesheets'));
});
