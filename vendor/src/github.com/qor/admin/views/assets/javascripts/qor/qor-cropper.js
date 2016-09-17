(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define(['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  var URL = window.URL || window.webkitURL;
  var NAMESPACE = 'qor.cropper';

  // Events
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CHANGE = 'change.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_SHOWN = 'shown.qor.modal';
  var EVENT_HIDDEN = 'hidden.qor.modal';

  // Classes
  var CLASS_TOGGLE = '.qor-cropper__toggle';
  var CLASS_CANVAS = '.qor-cropper__canvas';
  var CLASS_WRAPPER = '.qor-cropper__wrapper';
  var CLASS_OPTIONS = '.qor-cropper__options';
  var CLASS_SAVE = '.qor-cropper__save';
  var CLASS_DELETE = '.qor-cropper__toggle--delete';
  var CLASS_CROP = '.qor-cropper__toggle--crop';
  var CLASS_UNDO = '.qor-fieldset__undo';

  function capitalize(str) {
    if (typeof str === 'string') {
      str = str.charAt(0).toUpperCase() + str.substr(1);
    }

    return str;
  }

  function getLowerCaseKeyObject(obj) {
    var newObj = {};
    var key;

    if ($.isPlainObject(obj)) {
      for (key in obj) {
        if (obj.hasOwnProperty(key)) {
          newObj[String(key).toLowerCase()] = obj[key];
        }
      }
    }

    return newObj;
  }

  function getValueByNoCaseKey(obj, key) {
    var originalKey = String(key);
    var lowerCaseKey = originalKey.toLowerCase();
    var upperCaseKey = originalKey.toUpperCase();
    var capitalizeKey = capitalize(originalKey);

    if ($.isPlainObject(obj)) {
      return (obj[lowerCaseKey] || obj[capitalizeKey] || obj[upperCaseKey]);
    }
  }

  function replaceText(str, data) {
    if (typeof str === 'string') {
      if (typeof data === 'object') {
        $.each(data, function (key, val) {
          str = str.replace('${' + String(key).toLowerCase() + '}', val);
        });
      }
    }

    return str;
  }

  function QorCropper(element, options) {
    this.$element = $(element);
    this.options = $.extend(true, {}, QorCropper.DEFAULTS, $.isPlainObject(options) && options);
    this.data = null;
    this.init();
  }

  QorCropper.prototype = {
    constructor: QorCropper,

    init: function () {
      var options = this.options;
      var $this = this.$element;
      var $parent = $this.closest(options.parent);
      var $list;
      var data;
      var outputValue;
      var fetchUrl;
      var _this = this;
      var imageData;

      if (!$parent.length) {
        $parent = $this.parent();
      }

      this.$parent = $parent;
      this.$output = $parent.find(options.output);
      this.$list = $list = $parent.find(options.list);

      fetchUrl = this.$output.data().fetchSizedata;

      if (!$list.find('img').attr('src')) {
        $list.find('ul').hide();
      }

      if (fetchUrl) {
        $.getJSON(fetchUrl,function(data){
          imageData = JSON.parse(data.MediaOption);
          _this.$output.val(JSON.stringify(data));
          _this.data = imageData || {};
          _this.build();
          _this.bind();
        });
      } else {
        outputValue = $.trim(this.$output.val());
        if (outputValue) {
          data = JSON.parse(outputValue);
        }

        this.data = data || {};
        this.build();
        this.bind();
      }
    },

    build: function () {
      var textData = this.$output.data(),
          text = {
            title: textData.cropperTitle,
            ok: textData.cropperOk,
            cancel: textData.cropperCancel
          },
          replaceTexts = this.options.text;

      if (text.ok && text.title && text.cancel) {
        replaceTexts = text;
      }

      this.wrap();
      this.$modal = $(replaceText(QorCropper.MODAL, replaceTexts)).appendTo('body');
    },

    unbuild: function () {
      this.$modal.remove();
      this.unwrap();
    },

    wrap: function () {
      var $list = this.$list;
      var $img;

      $list.find('li').append(QorCropper.TOGGLE);
      $img = $list.find('img');
      $img.wrap(QorCropper.CANVAS);
      this.center($img);
    },

    unwrap: function () {
      var $list = this.$list;

      $list.find(CLASS_TOGGLE).remove();
      $list.find(CLASS_CANVAS).each(function () {
        var $this = $(this);

        $this.before($this.html()).remove();
      });
    },

    bind: function () {
      this.$element.
        on(EVENT_CHANGE, $.proxy(this.read, this));

      this.$list.
        on(EVENT_CLICK, $.proxy(this.click, this));

      this.$modal.
        on(EVENT_SHOWN, $.proxy(this.start, this)).
        on(EVENT_HIDDEN, $.proxy(this.stop, this));
    },

    unbind: function () {
      this.$element.
        off(EVENT_CHANGE, this.read);

      this.$list.
        off(EVENT_CLICK, this.click);

      this.$modal.
        off(EVENT_SHOWN, this.start).
        off(EVENT_HIDDEN, this.stop);
    },

    click: function (e) {
      var target = e.target;
      var $target;
      var data = this.data;
      var $alert;

      if (target === this.$list[0]) {
        return;
      }

      $target = $(target);

      if ($target.closest(CLASS_DELETE).size()){
        data.Delete = true;

        this.$output.val(JSON.stringify(data));
        this.$list.hide();

        $alert = $(QorCropper.ALERT);
        $alert.find(CLASS_UNDO).one(EVENT_CLICK, function () {
          $alert.remove();
          this.$list.show();
          delete data.Delete;
          this.$output.val(JSON.stringify(data));
        }.bind(this));
        this.$parent.find('.qor-fieldset').append($alert);
      }

      if ($target.closest(CLASS_CROP).size()) {
        $target = $target.closest('li').find('img');
        this.$target = $target;
        this.$modal.qorModal('show');
      }
    },

    read: function (e) {
      var files = e.target.files;
      var file;

      if (files && files.length) {
        file = files[0];

        if (/^image\/\w+$/.test(file.type) && URL) {
          this.load(URL.createObjectURL(file));
        } else {
          this.$list.empty().html(QorCropper.FILE_LIST.replace('{{filename}}', file.name));
        }
      }
    },

    load: function (url, callback) {
      var options = this.options;
      var _this = this;
      var $list = this.$list;
      var $ul = $list.find('ul');
      var data = this.data || {};
      var $image;
      var imageLength;

      if (!$ul.length || !$ul.find('li').length) {
        $ul  = $(QorCropper.LIST);
        $list.html($ul);
        this.wrap();
      }

      $ul.show(); // show ul when it is hidden

      $image = $list.find('img');
      imageLength = $image.size();
      $image.one('load', function () {
        var $this = $(this);
        var naturalWidth = this.naturalWidth;
        var naturalHeight = this.naturalHeight;
        var sizeData = $this.data();
        var sizeResolution = sizeData.sizeResolution;
        var sizeName = sizeData.sizeName;
        var emulateImageData = {};
        var emulateCropData = {};
        var aspectRatio;
        var width = sizeData.sizeResolutionWidth;
        var height = sizeData.sizeResolutionHeight;

        if (sizeResolution) {
          if (!width && !height) {
            width = getValueByNoCaseKey(sizeResolution, 'width');
            height = getValueByNoCaseKey(sizeResolution, 'height');
          }
          aspectRatio = width / height;

          if (naturalHeight * aspectRatio > naturalWidth) {
            width = naturalWidth;
            height = width / aspectRatio;
          } else {
            height = naturalHeight;
            width = height * aspectRatio;
          }

          emulateImageData = {
            naturalWidth: naturalWidth,
            naturalHeight: naturalHeight
          };

          emulateCropData = {
            x: Math.round((naturalWidth - width) / 2),
            y: Math.round((naturalHeight - height) / 2),
            width: Math.round(width),
            height: Math.round(height)
          };

          _this.preview($this, emulateImageData, emulateCropData);

          if (sizeName) {
            data.crop = true;

            if (!data[options.key]) {
              data[options.key] = {};
            }

            data[options.key][sizeName] = emulateCropData;
          }
        } else {
          _this.center($this);
        }

        _this.$output.val(JSON.stringify(data));

        // callback after load complete
        if (sizeName && Object.keys(data[options.key]).length >= imageLength) {
          if (callback && $.isFunction(callback)) {
            callback();
          }
        }
      }).attr('src', url).data('originalUrl', url);

      $list.show();
    },

    start: function () {
      var options = this.options;
      var $modal = this.$modal;
      var $target = this.$target;
      var sizeData = $target.data();
      var sizeName = sizeData.sizeName || 'original';
      var sizeResolution = sizeData.sizeResolution;
      var $clone = $('<img>').attr('src', sizeData.originalUrl);
      var data = this.data || {};
      var _this = this;
      var sizeAspectRatio = NaN;
      var sizeWidth = sizeData.sizeResolutionWidth;
      var sizeHeight = sizeData.sizeResolutionHeight;
      var list;

      if (sizeResolution) {
        if (!sizeWidth && !sizeHeight) {
          sizeWidth = getValueByNoCaseKey(sizeResolution, 'width');
          sizeHeight = getValueByNoCaseKey(sizeResolution, 'height');
        }
        sizeAspectRatio = sizeWidth / sizeHeight;
      }

      if (!data[options.key]) {
        data[options.key] = {};
      }

      $modal.trigger('enable.qor.material').find(CLASS_WRAPPER).html($clone);

      list = this.getList(sizeAspectRatio);

      if (list) {
        $modal.find(CLASS_OPTIONS).show().append(list);
      }

      $clone.cropper({
        aspectRatio: sizeAspectRatio,
        data: getLowerCaseKeyObject(data[options.key][sizeName]),
        background: false,
        movable: false,
        zoomable: false,
        scalable: false,
        rotatable: false,
        checkImageOrigin: false,
        autoCropArea: 1,

        built: function () {
          $modal.find(CLASS_SAVE).one(EVENT_CLICK, function () {
            var cropData = $clone.cropper('getData', true);
            var syncData = [];
            var url;
            
            data.crop = true;
            data[options.key][sizeName] = cropData;
            _this.imageData = $clone.cropper('getImageData');
            _this.cropData = cropData;

            url = $clone.cropper('getCroppedCanvas').toDataURL();

            $modal.find(CLASS_OPTIONS + ' input').each(function () {
              var $this = $(this);

              if ($this.prop('checked')) {
                syncData.push($this.attr('name'));
              }
            });

            _this.output(url, syncData);
            $modal.qorModal('hide');
          });
        }
      });
    },

    stop: function () {
      this.$modal.
        trigger('disable.qor.material').
        find(CLASS_WRAPPER + ' > img').
          cropper('destroy').
          remove().
          end().
        find(CLASS_OPTIONS).
          hide().
          find('ul').
            remove();
    },

    getList: function (aspectRatio) {
      var list = [];

      this.$list.find('img').not(this.$target).each(function () {
        var data = $(this).data();
        var resolution = data.sizeResolution;
        var name = data.sizeName;
        var width = data.sizeResolutionWidth;
        var height = data.sizeResolutionHeight;

        if (resolution) {
          if (!width && !height) {
            width = getValueByNoCaseKey(resolution, 'width');
            height = getValueByNoCaseKey(resolution, 'height');
          }

          if (width / height === aspectRatio) {
            list.push(
              '<label>' +
                '<input type="checkbox" name="' + name + '" checked> ' +
                '<span>' + name +
                  '<small>(' + width + '&times;' + height + ' px)</small>' +
                '</span>' +
              '</label>'
            );
          }
        }
      });

      return list.length ? ('<ul><li>' + list.join('</li><li>') + '</li></ul>') : '';
    },

    output: function (url, data) {
      var $target = this.$target;

      if (url) {
        this.center($target.attr('src', url), true);
      } else {
        this.preview($target);
      }

      if ($.isArray(data) && data.length) {
        this.autoCrop(url, data);
      }

      this.$output.val(JSON.stringify(this.data)).trigger(EVENT_CHANGE);
    },

    preview: function ($target, emulateImageData, emulateCropData) {
      var $canvas = $target.parent();
      var $container = $canvas.parent();
      var containerWidth = $container.width();
      var containerHeight = $container.height();
      var imageData = emulateImageData || this.imageData;
      var cropData = $.extend({}, emulateCropData || this.cropData); // Clone one to avoid changing it
      var aspectRatio = cropData.width / cropData.height;
      var canvasWidth = containerWidth;
      var scaledRatio;

      if (canvasWidth == 0 || imageData.naturalWidth == 0 || imageData.naturalHeight == 0) {
        return;
      }

      if (containerHeight * aspectRatio <= containerWidth) {
        canvasWidth = containerHeight * aspectRatio;
      }

      scaledRatio = cropData.width / canvasWidth;

      $target.css({
        maxWidth: imageData.naturalWidth / scaledRatio,
        maxHeight: imageData.naturalHeight / scaledRatio
      });

      this.center($target);
    },

    center: function ($target, reset) {
      $target.each(function () {
        var $this = $(this);
        var $canvas = $this.parent();
        var $container = $canvas.parent();

        function center() {
          var containerHeight = $container.height();
          var canvasHeight = $canvas.height();
          var marginTop = 'auto';

          if (canvasHeight < containerHeight) {
            marginTop = (containerHeight - canvasHeight) / 2;
          }

          $canvas.css('margin-top', marginTop);
        }

        if (reset) {
          $canvas.add($this).removeAttr('style');
        }

        if (this.complete) {
          center.call(this);
        } else {
          this.onload = center;
        }
      });
    },

    autoCrop: function (url, data) {
      var cropData = this.cropData;
      var cropOptions = this.data[this.options.key];
      var _this = this;

      this.$list.find('img').not(this.$target).each(function () {
        var $this = $(this);
        var sizeName = $this.data('sizeName');

        if ($.inArray(sizeName, data) > -1) {
          cropOptions[sizeName] = $.extend({}, cropData);

          if (url) {
            _this.center($this.attr('src', url), true);
          } else {
            _this.preview($this);
          }
        }
      });
    },

    destroy: function () {
      this.unbind();
      this.unbuild();
      this.$element.removeData(NAMESPACE);
    }
  };

  QorCropper.DEFAULTS = {
    parent: false,
    output: false,
    list: false,
    key: 'data',
    data: null,
    text: {
      title: 'Crop the image',
      ok: 'OK',
      cancel: 'Cancel'
    }
  };

  QorCropper.TOGGLE = ('<div class="qor-cropper__toggle">' +
      '<div class="qor-cropper__toggle--crop"><i class="material-icons">crop</i></div>' +
      '<div class="qor-cropper__toggle--delete"><i class="material-icons">delete</i></div>' +
    '</div>'
  );

  QorCropper.ALERT = (
    '<div class="qor-fieldset__alert">' +
      '<button class="mdl-button mdl-button--accent mdl-js-button mdl-js-ripple-effect qor-fieldset__undo" type="button">Undo delete</button>' +
    '</div>'
  );

  QorCropper.CANVAS = '<div class="qor-cropper__canvas"></div>';
  QorCropper.LIST = '<ul><li><img></li></ul>';
  QorCropper.FILE_LIST = '<div class="qor-file__list-item"><span><span>{{filename}}</span></span>';
  QorCropper.MODAL = (
    '<div class="qor-modal fade" tabindex="-1" role="dialog" aria-hidden="true">' +
      '<div class="mdl-card mdl-shadow--2dp" role="document">' +
        '<div class="mdl-card__title">' +
          '<h2 class="mdl-card__title-text">${title}</h2>' +
        '</div>' +
        '<div class="mdl-card__supporting-text">' +
          '<div class="qor-cropper__wrapper"></div>' +
          '<div class="qor-cropper__options">' +
            '<p>Sync cropping result to:</p>' +
          '</div>' +
        '</div>' +
        '<div class="mdl-card__actions mdl-card--border">' +
          '<a class="mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect qor-cropper__save">${ok}</a>' +
          '<a class="mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect" data-dismiss="modal">${cancel}</a>' +
        '</div>' +
        '<div class="mdl-card__menu">' +
          '<button class="mdl-button mdl-button--icon mdl-js-button mdl-js-ripple-effect" data-dismiss="modal" aria-label="close">' +
            '<i class="material-icons">close</i>' +
          '</button>' +
        '</div>' +
      '</div>' +
    '</div>'
  );

  QorCropper.plugin = function (option) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var options;
      var fn;

      if (!data) {
        if (!$.fn.cropper) {
          return;
        }

        if (/destroy/.test(option)) {
          return;
        }

        options = $.extend(true, {}, $this.data(), typeof option === 'object' && option);
        $this.data(NAMESPACE, (data = new QorCropper(this, options)));
      }

      if (typeof option === 'string' && $.isFunction(fn = data[option])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-file__input';
    var options = {
          parent: '.qor-file',
          output: '.qor-file__options',
          list: '.qor-file__list',
          key: 'CropOptions'
        };

    $(document).
      on(EVENT_ENABLE, function (e) {
        QorCropper.plugin.call($(selector, e.target), options);
      }).
      on(EVENT_DISABLE, function (e) {
        QorCropper.plugin.call($(selector, e.target), 'destroy');
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorCropper;

});
