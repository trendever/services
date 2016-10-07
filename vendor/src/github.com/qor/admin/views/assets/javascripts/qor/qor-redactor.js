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

  var $window = $(window);
  var NAMESPACE = 'qor.redactor';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_FOCUS = 'focus.' + NAMESPACE;
  var EVENT_BLUR = 'blur.' + NAMESPACE;
  var EVENT_IMAGE_UPLOAD = 'imageupload.' + NAMESPACE;
  var EVENT_IMAGE_DELETE = 'imagedelete.' + NAMESPACE;
  var EVENT_SHOWN = 'shown.qor.modal';
  var EVENT_HIDDEN = 'hidden.qor.modal';

  var CLASS_WRAPPER = '.qor-cropper__wrapper';
  var CLASS_SAVE = '.qor-cropper__save';

  function encodeCropData(data) {
    var nums = [];

    if ($.isPlainObject(data)) {
      $.each(data, function () {
        nums.push(arguments[1]);
      });
    }

    return nums.join();
  }

  function decodeCropData(data) {
    var nums = data && data.split(',');

    data = null;

    if (nums && nums.length === 4) {
      data = {
        x: Number(nums[0]),
        y: Number(nums[1]),
        width: Number(nums[2]),
        height: Number(nums[3])
      };
    }

    return data;
  }

  function capitalize (str) {
    if (typeof str === 'string') {
      str = str.charAt(0).toUpperCase() + str.substr(1);
    }

    return str;
  }

  function getCapitalizeKeyObject (obj) {
    var newObj = {},
        key;

    if ($.isPlainObject(obj)) {
      for (key in obj) {
        if (obj.hasOwnProperty(key)) {
          newObj[capitalize(key)] = obj[key];
        }
      }
    }

    return newObj;
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

  function QorRedactor(element, options) {
    this.$element = $(element);
    this.options = $.extend(true, {}, QorRedactor.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorRedactor.prototype = {
    constructor: QorRedactor,

    init: function () {
      var options = this.options;
      var $this = this.$element;
      var $parent = $this.closest(options.parent);

      if (!$parent.length) {
        $parent = $this.parent();
      }

      this.$parent = $parent;
      this.$button = $(QorRedactor.BUTTON);
      this.$modal = $(replaceText(QorRedactor.MODAL, options.text)).appendTo('body');
      this.bind();
    },

    bind: function () {
      var $parent = this.$parent;
      var click = $.proxy(this.click, this);

      this.$element.
        on(EVENT_IMAGE_UPLOAD, function (e, image) {
          $(image).on(EVENT_CLICK, click);
        }).
        on(EVENT_IMAGE_DELETE, function (e, image) {
          $(image).off(EVENT_CLICK, click);
        }).
        on(EVENT_FOCUS, function () {
          $parent.find('img').off(EVENT_CLICK, click).on(EVENT_CLICK, click);
        }).
        on(EVENT_BLUR, function () {
          $parent.find('img').off(EVENT_CLICK, click);
        });

      $window.on(EVENT_CLICK, $.proxy(this.removeButton, this));
    },

    unbind: function () {
      this.$element.
        off(EVENT_IMAGE_UPLOAD).
        off(EVENT_IMAGE_DELETE).
        off(EVENT_FOCUS).
        off(EVENT_BLUR);

      $window.off(EVENT_CLICK, this.removeButton);
    },

    click: function (e) {
      e.stopPropagation();
      setTimeout($.proxy(this.addButton, this, $(e.target)), 1);
    },

    addButton: function ($image) {
      this.$button.
        prependTo($image.parent()).
        off(EVENT_CLICK).
        one(EVENT_CLICK, $.proxy(this.crop, this, $image));
    },

    removeButton: function () {
      this.$button.off(EVENT_CLICK).detach();
    },

    crop: function ($image) {
      var options = this.options;
      var url = $image.attr('src');
      var originalUrl = url;
      var $clone = $('<img>');
      var $modal = this.$modal;

      if ($.isFunction(options.replace)) {
        originalUrl = options.replace(originalUrl);
      }

      $clone.attr('src', originalUrl);
      $modal.one(EVENT_SHOWN, function () {
        $clone.cropper({
          data: decodeCropData($image.attr('data-crop-options')),
          background: false,
          movable: false,
          zoomable: false,
          scalable: false,
          rotatable: false,
          checkImageOrigin: false,

          built: function () {
            $modal.find(CLASS_SAVE).one(EVENT_CLICK, function () {
              var cropData = $clone.cropper('getData', true);

              $.ajax(options.remote, {
                type: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({
                  Url: url,
                  CropOptions: {
                    original: getCapitalizeKeyObject(cropData)
                  },
                  Crop: true
                }),
                dataType: 'json',

                success: function (response) {
                  if ($.isPlainObject(response) && response.url) {
                    $image.attr('src', response.url).attr('data-crop-options', encodeCropData(cropData)).removeAttr('style').removeAttr('rel');

                    if ($.isFunction(options.complete)) {
                      options.complete();
                    }

                    $modal.qorModal('hide');
                  }
                }
              });
            });
          }
        });
      }).one(EVENT_HIDDEN, function () {
        $clone.cropper('destroy').remove();
      }).qorModal('show').find(CLASS_WRAPPER).append($clone);
    },

    destroy: function () {
      this.unbind();
      this.$modal.qorModal('hide').remove();
      this.$element.removeData(NAMESPACE);
    }
  };

  QorRedactor.DEFAULTS = {
    remote: false,
    parent: false,
    toggle: false,
    replace: null,
    complete: null,
    text: {
      title: 'Crop the image',
      ok: 'OK',
      cancel: 'Cancel'
    }
  };

  QorRedactor.BUTTON = '<span class="qor-cropper__toggle--redactor" contenteditable="false">Crop</span>';
  QorRedactor.MODAL = (
    '<div class="qor-modal fade" tabindex="-1" role="dialog" aria-hidden="true">' +
      '<div class="mdl-card mdl-shadow--2dp" role="document">' +
        '<div class="mdl-card__title">' +
          '<h2 class="mdl-card__title-text">${title}</h2>' +
        '</div>' +
        '<div class="mdl-card__supporting-text">' +
          '<div class="qor-cropper__wrapper"></div>' +
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

  QorRedactor.plugin = function (option) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var config;
      var fn;

      if (!data) {
        if (!$.fn.redactor) {
          return;
        }

        if (/destroy/.test(option)) {
          return;
        }

        $this.data(NAMESPACE, (data = {}));
        config = {
          imageUpload: $this.data("uploadUrl"),
          fileUpload: $this.data("uploadUrl"),

          initCallback: function () {
            if (!$this.data("cropUrl")) {
              return;
            }

            $this.data(NAMESPACE, (data = new QorRedactor($this, {
              remote: $this.data("cropUrl"),
              text: $this.data("text"),
              parent: '.qor-field',
              toggle: '.qor-cropper__toggle--redactor',
              replace: function (url) {
                return url.replace(/\.\w+$/, function (extension) {
                  return '.original' + extension;
                });
              },
              complete: $.proxy(function () {
                this.code.sync();
              }, this)
            })));
          },

          focusCallback: function (/*e*/) {
            $this.triggerHandler(EVENT_FOCUS);
          },

          blurCallback: function (/*e*/) {
            $this.triggerHandler(EVENT_BLUR);
          },

          imageUploadCallback: function (/*image, json*/) {
            $this.triggerHandler(EVENT_IMAGE_UPLOAD, arguments[0]);
          },

          imageDeleteCallback: function (/*url, image*/) {
            $this.triggerHandler(EVENT_IMAGE_DELETE, arguments[1]);
          }
        };

        $.extend(config, $this.data("redactorSettings"));
        $this.redactor(config);
      } else {
        if (/destroy/.test(option)) {
          $this.redactor('core.destroy');
        }
      }

      if (typeof option === 'string' && $.isFunction(fn = data[option])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = 'textarea[data-toggle="qor.redactor"]';

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorRedactor.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorRedactor.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorRedactor;

});
