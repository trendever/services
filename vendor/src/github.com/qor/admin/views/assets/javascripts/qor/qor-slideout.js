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

  var $document = $(document);
  var FormData = window.FormData;
  var _ = window._;
  var NAMESPACE = 'qor.slideout';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_KEYUP = 'keyup.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_SUBMIT = 'submit.' + NAMESPACE;
  var EVENT_SHOW = 'show.' + NAMESPACE;
  var EVENT_SHOWN = 'shown.' + NAMESPACE;
  var EVENT_HIDE = 'hide.' + NAMESPACE;
  var EVENT_HIDDEN = 'hidden.' + NAMESPACE;
  var EVENT_TRANSITIONEND = 'transitionend';
  var CLASS_OPEN = 'qor-slideout-open';
  var CLASS_IS_SHOWN = 'is-shown';
  var CLASS_IS_SLIDED = 'is-slided';
  var CLASS_IS_SELECTED = 'is-selected';
  var CLASS_MAIN_CONTENT = '.mdl-layout__content.qor-page';
  var CLASS_HEADER_LOCALE = '.qor-actions__locale';

  function QorSlideout(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSlideout.DEFAULTS, $.isPlainObject(options) && options);
    this.slided = false;
    this.disabled = false;
    this.init();
  }

  QorSlideout.prototype = {
    constructor: QorSlideout,

    init: function () {
      this.build();
      this.bind();
    },

    build: function () {
      var $slideout;

      this.$slideout = $slideout = $(QorSlideout.TEMPLATE).appendTo('body');
      this.$title = $slideout.find('.qor-slideout__title');
      this.$body = $slideout.find('.qor-slideout__body');
      this.$bodyClass = $('body').prop('class');
    },

    unbuild: function () {
      this.$title = null;
      this.$body = null;
      this.$slideout.remove();
    },

    bind: function () {
      this.$slideout.
        on(EVENT_SUBMIT, 'form', $.proxy(this.submit, this));

      $document.
        on(EVENT_KEYUP, $.proxy(this.keyup, this)).
        on(EVENT_CLICK, $.proxy(this.click, this));
    },

    unbind: function () {
      this.$slideout.
        off(EVENT_SUBMIT, this.submit);

      $document.
        off(EVENT_KEYUP, this.keyup).
        off(EVENT_CLICK, this.click);
    },

    keyup: function (e) {
      if (e.which === 27) {
        this.hide();
        this.removeSelectedClass();
      }
    },

    loadScript: function (src, url, response) {
      var options = this.options;
      var script = document.createElement('script');
      script.src = src;
      script.onload = function () {
        // exec qorSliderAfterShow after script loaded
        if (options.afterShow){
          var qorSliderAfterShow = $.fn.qorSliderAfterShow;
          for (var name in qorSliderAfterShow) {
            if (qorSliderAfterShow.hasOwnProperty(name)) {
              qorSliderAfterShow[name].call(this, url, response);
            }
          }
        }
      };
      document.body.appendChild(script);
    },

    loadStyle: function (src) {
      var ss = document.createElement('link');
      ss.type = 'text/css';
      ss.rel = 'stylesheet';
      ss.href = src;
      document.getElementsByTagName('head')[0].appendChild(ss);
    },

    pushArrary: function ($ele,prop) {
      var array = [];
      $ele.each(function () {
        array.push($(this).prop(prop));
      });
      return array;
    },

    removeSelectedClass: function () {
      this.$element.find('tbody > tr[data-url],a[data-url]').removeClass(CLASS_IS_SELECTED);
    },

    click: function (e) {
      var $this = this.$element;
      var slideout = this.$slideout.get(0);
      var target = e.target;
      var $target;
      var data;

      function toggleClass() {
        $this.find('tbody > tr[data-url],a[data-url]').removeClass(CLASS_IS_SELECTED);
        $target.addClass(CLASS_IS_SELECTED);
      }

      if (e.isDefaultPrevented()) {
        return;
      }

      while (target !== document) {
        $target = $(target);

        if ($target.prop('disabled')) {
          break;
        }

        if (target === slideout) {
          break;
        } else if ($target.data('dismiss') === 'slideout') {
          this.hide();
          this.removeSelectedClass();
          break;
        } else if ($target.is('table.qor-table > tbody > tr[data-url]')) {
          if ($(e.target).parents('.qor-table__actions').size() > 0) {
            return;
          }

          if ($target.hasClass(CLASS_IS_SELECTED)) {
            this.hide();
            this.removeSelectedClass();
          } else {
            toggleClass();
            data = $target.data();
            this.load(data.url);
          }

          break;
        } else if ($target.data('url')) {
          e.preventDefault();

          if ($target.hasClass(CLASS_IS_SELECTED)) {
            this.hide();
            this.removeSelectedClass();
          } else {
            toggleClass();
            data = $target.data();
            this.load(data.url, data);
          }
          break;
        } else {
          if ($target.is('a')) {
            break;
          }

          if (target) {
            target = target.parentNode;
          } else {
            break;
          }
        }
      }
    },

    submit: function (e) {
      var $slideout = this.$slideout;
      var $body = this.$body;
      var form = e.target;
      var $form = $(form);
      var _this = this;
      var $submit = $form.find(':submit');

      if (FormData) {
        e.preventDefault();

        $.ajax($form.prop('action'), {
          method: $form.prop('method'),
          data: new FormData(form),
          dataType: 'html',
          processData: false,
          contentType: false,
          beforeSend: function () {
            $submit.prop('disabled', true);
          },
          success: function (html) {
            var returnUrl = $form.data('returnUrl');

            if (returnUrl == 'refresh') {
              _this.refresh();
              return;
            }

            if (returnUrl && returnUrl != 'refresh') {
              _this.load(returnUrl);
            } else {
              var prefix = '/' + location.pathname.split('/')[1];
              var flashStructs = [];
              $(html).find('.qor-alert').each(function (i, e) {
                var message = $(e).find('.qor-alert-message').text().trim();
                var type = $(e).data('type');
                if (message !== '') {
                  flashStructs.push({ Type: type, Message: message, Keep: true });
                }
              });
              if (flashStructs.length > 0) {
                document.cookie = 'qor-flashes=' + btoa(JSON.stringify(flashStructs)) + '; path=' + prefix;
              }
              _this.refresh();
            }
          },
          error: function (xhr, textStatus, errorThrown) {
            var $error;

            // Custom HTTP status code
            if (xhr.status === 422) {

              // Clear old errors
              $body.find('.qor-error').remove();
              $form.find('.qor-field').removeClass('is-error').find('.qor-field__error').remove();

              // Append new errors
              $error = $(xhr.responseText).find('.qor-error');
              $form.before($error);

              $error.find('> li > label').each(function () {
                var $label = $(this);
                var id = $label.attr('for');

                if (id) {
                  $form.find('#' + id).
                    closest('.qor-field').
                    addClass('is-error').
                    append($label.clone().addClass('qor-field__error'));
                }
              });

              // Scroll to top to view the errors
              $slideout.scrollTop(0);
            } else {
              window.alert([textStatus, errorThrown].join(': '));
            }
          },
          complete: function () {
            $submit.prop('disabled', false);
          }
        });
      }
    },

    load: function (url, data) {
      var options = this.options;
      var method;
      var dataType;
      var load;

      if (!url) {
        return;
      }

      data = $.isPlainObject(data) ? data : {};

      method = data.method ? data.method : 'GET';
      dataType = data.datatype ? data.datatype : 'html';

      data.url = data.method = data.datatype = data.ajaxForm = data.upgraded = undefined;

      load = $.proxy(function () {
        $.ajax(url, {
          method: method,
          data: data,
          dataType: dataType,
          success: $.proxy(function (response) {
            var $response;
            var $content;

            if (method === 'GET') {
              $response = $(response);

              $content = $response.find(CLASS_MAIN_CONTENT);

              if (!$content.length) {
                return;
              }

              // Get response body tag: http://stackoverflow.com/questions/7001926/cannot-get-body-element-from-ajax-response
              var dataBody = response.match(/<\s*body.*>[\s\S]*<\s*\/body\s*>/ig);
              // if no body tag return
              if (dataBody) {

                dataBody  = dataBody.join('');
                dataBody  = dataBody.replace(/<\s*body/gi, '<div');
                dataBody  = dataBody.replace(/<\s*\/body/gi, '</div');
                var bodyClass = $(dataBody).prop('class');
                $('body').removeClass().addClass(bodyClass);

                // Get links and scripts, compare slideout and inline, load style and script if has new style or script.
                var $slideoutStyles = $response.filter('link');
                var $currentPageStyles = $('link');
                var $slideoutScripts = $response.filter('script');
                var $currentPageScripts = $('script');

                var slideoutStylesUrls = this.pushArrary($slideoutStyles, 'href');
                var currentPageStylesUrls = this.pushArrary($currentPageStyles, 'href');

                var slideoutScriptsUrls = this.pushArrary($slideoutScripts, 'src');
                var currentPageScriptsUrls = this.pushArrary($currentPageScripts, 'src');

                var styleDifferenceUrl  = _.difference(slideoutStylesUrls, currentPageStylesUrls);
                var scriptDifferenceUrl = _.difference(slideoutScriptsUrls, currentPageScriptsUrls);

                var styleDifferenceUrlLength = styleDifferenceUrl.length;
                var scriptDifferenceUrlLength = scriptDifferenceUrl.length;

                if (styleDifferenceUrlLength === 1){
                  this.loadStyle(styleDifferenceUrl);
                } else if (styleDifferenceUrlLength > 1){
                  for (var i = styleDifferenceUrlLength - 1; i >= 0; i--) {
                    this.loadStyle(styleDifferenceUrl[i]);
                  }
                }

                if (scriptDifferenceUrlLength === 1){
                  this.loadScript(scriptDifferenceUrl, url, response);
                } else if (scriptDifferenceUrlLength > 1){
                  for (var j = scriptDifferenceUrlLength - 1; j >= 0; j--) {
                    this.loadScript(scriptDifferenceUrl[j], url, response);
                  }
                }

              }

              // end

              $content.find('.qor-button--cancel').attr('data-dismiss', 'slideout').removeAttr('href');
              this.$title.html($response.find(options.title).html());
              this.$body.html($content.html());
              this.$body.find(CLASS_HEADER_LOCALE).remove();

              this.$slideout.one(EVENT_SHOWN, function () {

                // Enable all Qor components within the slideout
                $(this).trigger('enable');
              }).one(EVENT_HIDDEN, function () {

                // Destroy all Qor components within the slideout
                $(this).trigger('disable');

              });

              this.show();

              $('[data-toggle="qor.datetimepicker"]').materialDatePicker({ format : 'YYYY-MM-DD HH:mm' });
              $('[data-toggle="qor.datepicker"]').materialDatePicker({ format : 'YYYY-MM-DD', time: false });

              // callback for after slider loaded HTML
              if (options.afterShow){
                var qorSliderAfterShow = $.fn.qorSliderAfterShow;

                for (var name in qorSliderAfterShow) {
                  if (qorSliderAfterShow.hasOwnProperty(name)) {
                    qorSliderAfterShow[name].call(this, url, response);
                  }
                }

              }

            } else {
              if (data.returnUrl) {
                this.load(data.returnUrl);
              } else {
                this.refresh();
              }
            }


          }, this),


          error: $.proxy (function (response) {
            var errors;
            if ($('.qor-error span').size() > 0) {
              errors = $('.qor-error span').map(function () {
                return $(this).text();
              }).get().join(', ');
            } else {
              errors = response.responseText;
            }
            window.alert(errors);
          }, this)

        });
      }, this);

      if (this.slided) {
        this.hide();
        this.$slideout.one(EVENT_HIDDEN, load);
      } else {
        load();
      }
    },

    show: function () {
      var $slideout = this.$slideout;
      var showEvent;

      if (this.slided) {
        return;
      }

      showEvent = $.Event(EVENT_SHOW);
      $slideout.trigger(showEvent);

      if (showEvent.isDefaultPrevented()) {
        return;
      }

      $slideout.addClass(CLASS_IS_SHOWN).get(0).offsetWidth;
      $slideout.
        one(EVENT_TRANSITIONEND, $.proxy(this.shown, this)).
        addClass(CLASS_IS_SLIDED).
        scrollTop(0);
    },

    shown: function () {
      this.slided = true;

      // Disable to scroll body element
      $('body').addClass(CLASS_OPEN);

      this.$slideout.trigger(EVENT_SHOWN);
    },

    hide: function () {
      var $slideout = this.$slideout;
      var hideEvent;
      var $datePicker = $('.qor-datepicker').not('.hidden');

      $('body').removeClass().addClass(this.$bodyClass);

      if ($datePicker.size()){
        $datePicker.addClass('hidden');
      }

      if (!this.slided) {
        return;
      }

      hideEvent = $.Event(EVENT_HIDE);
      $slideout.trigger(hideEvent);

      if (hideEvent.isDefaultPrevented()) {
        return;
      }

      // empty body html when hide slideout
      this.$body.html('');

      $slideout.
        one(EVENT_TRANSITIONEND, $.proxy(this.hidden, this)).
        removeClass(CLASS_IS_SLIDED);
    },

    hidden: function () {
      this.slided = false;

      // Enable to scroll body element
      $('body').removeClass(CLASS_OPEN);

      this.$slideout.removeClass(CLASS_IS_SHOWN).trigger(EVENT_HIDDEN);
    },

    refresh: function () {
      this.hide();

      setTimeout(function () {
        window.location.reload();
      }, 350);
    },

    destroy: function () {
      this.unbind();
      this.unbuild();
      this.$element.removeData(NAMESPACE);
    }
  };

  QorSlideout.DEFAULTS = {
    title: false,
    content: false
  };

  QorSlideout.TEMPLATE = (
    '<div class="qor-slideout">' +
      '<div class="qor-slideout__header">' +
        '<button type="button" class="mdl-button mdl-button--icon mdl-js-button mdl-js-repple-effect qor-slideout__close" data-dismiss="slideout">' +
          '<span class="material-icons">close</span>' +
        '</button>' +
        '<h3 class="qor-slideout__title"></h3>' +
      '</div>' +
      '<div class="qor-slideout__body"></div>' +
    '</div>'
  );

  QorSlideout.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorSlideout(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-theme-slideout';
    var options = {
          title: '.qor-form-title, .mdl-layout-title',
          afterShow: $.fn.qorSliderAfterShow ? $.fn.qorSliderAfterShow : null
        };

    $(document).
      on(EVENT_ENABLE, function (e) {

        if (/slideout/.test(e.namespace)) {
          QorSlideout.plugin.call($(selector, e.target), options);
        }
      }).
      on(EVENT_DISABLE, function (e) {

        if (/slideout/.test(e.namespace)) {
          QorSlideout.plugin.call($(selector, e.target), 'destroy');
        }
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorSlideout;

});
