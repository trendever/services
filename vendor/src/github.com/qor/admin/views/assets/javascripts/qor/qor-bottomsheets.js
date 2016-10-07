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

  var FormData = window.FormData;
  var NAMESPACE = 'qor.bottomsheets';
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_SUBMIT = 'submit.' + NAMESPACE;
  var EVENT_RELOAD = 'reload.' + NAMESPACE;
  var EVENT_HIDE = 'hide.' + NAMESPACE;
  var EVENT_HIDDEN = 'hidden.' + NAMESPACE;
  var EVENT_KEYUP = 'keyup.' + NAMESPACE;
  var CLASS_OPEN = 'qor-bottomsheets-open';
  var CLASS_IS_SHOWN = 'is-shown';
  var CLASS_IS_SLIDED = 'is-slided';
  var CLASS_MAIN_CONTENT = '.mdl-layout__content.qor-page';
  var CLASS_BODY_CONTENT = '.qor-page__body';
  var CLASS_BODY_HEAD = '.qor-page__header';
  var CLASS_BOTTOMSHEETS = '.qor-bottomsheets';
  var CLASS_BOTTOMSHEETS_FILTER = '.qor-bottomsheet__filter';
  var CLASS_BOTTOMSHEETS_BUTTON = '.qor-bottomsheets__search-button';
  var CLASS_BOTTOMSHEETS_INPUT = '.qor-bottomsheets__search-input';
  var URL_GETQOR = 'http://www.getqor.com/';

  function QorBottomSheets(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorBottomSheets.DEFAULTS, $.isPlainObject(options) && options);
    this.disabled = false;
    this.resourseData = {};
    this.init();
  }

  function getUrlParameter(name, search) {
    name = name.replace(/[\[]/, '\\[').replace(/[\]]/, '\\]');
    var regex = new RegExp('[\\?&]' + name + '=([^&#]*)');
    var results = regex.exec(search);
    return results === null ? '' : decodeURIComponent(results[1].replace(/\+/g, ' '));
  }

  function updateQueryStringParameter(uri, key, value) {
    var escapedkey = String(key).replace(/[\\^$*+?.()|[\]{}]/g, '\\$&');
    var re = new RegExp('([?&])' + escapedkey + '=.*?(&|$)', 'i');
    var separator = uri.indexOf('?') !== -1 ? '&' : '?';
    if (uri.match(re)) {
      return uri.replace(re, '$1' + key + '=' + value + '$2');
    } else {
      return uri + separator + key + '=' + value;
    }
  }

  QorBottomSheets.prototype = {
    constructor: QorBottomSheets,

    init: function () {
      this.build();
      this.bind();
    },

    build: function () {
      var $bottomsheets = $(CLASS_BOTTOMSHEETS);

      if ($bottomsheets.size()) {
        $bottomsheets.remove();
      }

      this.$bottomsheets = $bottomsheets = $(QorBottomSheets.TEMPLATE).appendTo('body');
      this.$body = $bottomsheets.find('.qor-bottomsheets__body');
      this.$title = $bottomsheets.find('.qor-bottomsheets__title');
      this.$header = $bottomsheets.find('.qor-bottomsheets__header');
      this.$bodyClass = $('body').prop('class');
      this.filterURL = '';
      this.searchParams = '';

    },

    unbuild: function () {
      this.$body = null;
      this.$bottomsheets.remove();
    },

    bind: function () {
      this.$bottomsheets
        .on(EVENT_SUBMIT, 'form', this.submit.bind(this))
        .on(EVENT_CLICK, '[data-dismiss="bottomsheets"]', this.hide.bind(this))
        .on(EVENT_CLICK, '.qor-pagination a', this.pagination.bind(this))
        .on(EVENT_CLICK, CLASS_BOTTOMSHEETS_BUTTON, this.search.bind(this))
        .on(EVENT_KEYUP, this.keyup.bind(this))
        .on('selectorChanged.qor.selector', this.selectorChanged.bind(this))
        .on('filterChanged.qor.filter', this.filterChanged.bind(this));
    },

    unbind: function () {
      this.$bottomsheets
        .off(EVENT_SUBMIT, 'form', this.submit.bind(this))
        .off(EVENT_CLICK, '[data-dismiss="bottomsheets"]', this.hide.bind(this))
        .off(EVENT_CLICK, '.qor-pagination a', this.pagination.bind(this))
        .off(EVENT_CLICK, CLASS_BOTTOMSHEETS_BUTTON, this.search.bind(this))
        .off('selectorChanged.qor.selector', this.selectorChanged.bind(this))
        .off('filterChanged.qor.filter', this.filterChanged.bind(this));
    },

    bindActionData: function (actiondData) {
      var $form = this.$body.find('[data-toggle="qor-action-slideout"]').find('form');
      for (var i = actiondData.length - 1; i >= 0; i--) {
        $form.prepend('<input type="hidden" name="primary_values[]" value="' + actiondData[i] + '" />');
      }
    },

    filterChanged: function (e, search, key) {
      // if this event triggered:
      // search: ?locale_mode=locale, ?filters[Color].Value=2
      // key: search param name: locale_mode

      var loadUrl;

      loadUrl = this.constructloadURL(search, key);
      loadUrl && this.reload(loadUrl);
      return false;
    },

    selectorChanged: function (e, url, key) {
      // if this event triggered:
      // url: /admin/!remote_data_searcher/products/Collections?locale=en-US
      // key: search param key: locale

      var loadUrl;

      loadUrl = this.constructloadURL(url, key);
      loadUrl && this.reload(loadUrl);
      return false;
    },

    keyup: function (e) {
      var searchInput = this.$bottomsheets.find(CLASS_BOTTOMSHEETS_INPUT);

      if (e.which === 13 && searchInput.length && searchInput.is(':focus')) {
        this.search();
      }
    },

    search: function () {
      var $bottomsheets = this.$bottomsheets,
          param = '?keyword=',
          baseUrl = $bottomsheets.data().url,
          searchValue = $.trim($bottomsheets.find(CLASS_BOTTOMSHEETS_INPUT).val()),
          url = baseUrl + param + searchValue;

      this.reload(url);
    },

    pagination: function (e) {
      var $ele = $(e.target),
          url = $ele.prop('href');
      if (url) {
        this.reload(url);
      }
      return false;
    },

    reload: function (url) {
      var $content = this.$bottomsheets.find(CLASS_BODY_CONTENT);

      this.addLoading($content);
      this.fetchPage(url);
    },

    fetchPage: function (url) {
      var $bottomsheets = this.$bottomsheets,
          _this = this;

      $.get(url, function (response) {
        var $response = $(response).find(CLASS_MAIN_CONTENT),
            $responseHeader = $response.find(CLASS_BODY_HEAD),
            $responseBody = $response.find(CLASS_BODY_CONTENT);

        if ($responseBody.length) {
          $bottomsheets.find(CLASS_BODY_CONTENT).html($responseBody.html());

          if ($responseHeader.length) {
            _this.$body.find(CLASS_BODY_HEAD).html($responseHeader.html()).trigger('enable');
            _this.addHeaderClass();
          }
          // will trigger this event(relaod.qor.bottomsheets) when bottomsheets reload complete: like pagination, filter, action etc.
          $bottomsheets.trigger(EVENT_RELOAD);
        } else {
          _this.reload(url);
        }
      }).fail(function() {
        window.alert( "server error, please try again later!" );
      });
    },

    constructloadURL: function (url, key) {
      var fakeURL,
          value,
          filterURL = this.filterURL,
          bindUrl = this.$bottomsheets.data().url;

      if (!filterURL) {
        if (bindUrl) {
          filterURL = bindUrl;
        } else {
          return;
        }
      }

      fakeURL = new URL(URL_GETQOR + url);
      value = getUrlParameter(key, fakeURL.search);
      filterURL = this.filterURL = updateQueryStringParameter(filterURL, key, value);

      return filterURL;
    },

    addHeaderClass: function () {
      this.$body.find(CLASS_BODY_HEAD).hide();
      if (this.$bottomsheets.find(CLASS_BODY_HEAD).children(CLASS_BOTTOMSHEETS_FILTER).length) {
        this.$body.addClass('has-header').find(CLASS_BODY_HEAD).show();
      }
    },

    addLoading: function ($element) {
      $element.html('');
      var $loading = $(QorBottomSheets.TEMPLATE_LOADING).appendTo($element);
      window.componentHandler.upgradeElement($loading.children()[0]);
    },

    submit: function (e) {

      // will ingore submit event if need handle with other submit event: like select one, many...
      if (this.resourseData.ingoreSubmit) {
        return;
      }

      var $bottomsheets = this.$bottomsheets;
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
          dataType: 'json',
          processData: false,
          contentType: false,
          beforeSend: function () {
            $submit.prop('disabled', true);
          },
          success: function () {
            var returnUrl = $form.data('returnUrl');
            var refreshUrl = $form.data('refreshUrl');

            if (refreshUrl) {
              window.location.href = refreshUrl;
              return;
            }

            if (returnUrl == 'refresh') {
              _this.refresh();
              return;
            }

            if (returnUrl && returnUrl != 'refresh') {
              _this.load(returnUrl);
            } else {
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
              $bottomsheets.scrollTop(0);
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

    load: function (url, data, callback) {
      var options = this.options;
      var method;
      var dataType;
      var load;
      var actionData = data.actionData;
      var selectModal = this.resourseData.selectModal;
      var hasSearch = selectModal && $('.qor-search-container').length;
      var ingoreSubmit = this.resourseData.ingoreSubmit;
      var $bottomsheets = this.$bottomsheets;
      var $header = this.$header;
      var $body = this.$body;

      if (!url) {
        return;
      }

      this.show();
      this.addLoading($body);

      this.filterURL = url;
      $body.removeClass('has-header has-hint');

      data = $.isPlainObject(data) ? data : {};

      method = data.method ? data.method : 'GET';
      dataType = data.datatype ? data.datatype : 'html';

      load = $.proxy(function () {
        $.ajax(url, {
          method: method,
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

              if (ingoreSubmit) {
                $content.find(CLASS_BODY_HEAD).remove();
              }

              $content.find('.qor-button--cancel').attr('data-dismiss', 'bottomsheets');

              $body.html($content.html());
              this.$title.html($response.find(options.title).html());



              if (selectModal) {
                $body.find('.qor-button--new').data('ingoreSubmit',true).data('selectId',this.resourseData.selectId);
                if (selectModal != 'one' && this.resourseData.maxItem != '1') {
                  $body.addClass('has-hint');
                }
              }

              $header.find('.qor-button--new').remove();
              this.$title.after($body.find('.qor-button--new'));

              if (hasSearch) {
                $header.find('.qor-bottomsheets__search').remove();
                $header.prepend(QorBottomSheets.TEMPLATE_SEARCH);
              }

              if (actionData && actionData.length) {
                this.bindActionData(actionData);
              }

              $bottomsheets.trigger('enable');

              $bottomsheets.one(EVENT_HIDDEN, function () {
                $(this).trigger('disable');
              });


              this.addHeaderClass();
              $bottomsheets.data(data);

              // handle after opened callback
              if (callback && $.isFunction(callback)) {
                callback();
              }

              // callback for after bottomSheets loaded HTML
              // if (options.afterShow){
              //   var qorBottomsheetsAfterShow = $.fn.qorBottomsheetsAfterShow;

              //   for (var name in qorBottomsheetsAfterShow) {
              //     if (qorBottomsheetsAfterShow.hasOwnProperty(name) && $.isFunction(qorBottomsheetsAfterShow[name])) {
              //       qorBottomsheetsAfterShow[name].call(this, url, response);
              //     }
              //   }

              // }

            } else {
              if (data.returnUrl) {
                this.load(data.returnUrl);
              } else {
                this.refresh();
              }
            }


          }, this),


          error: $.proxy (function (response) {
            this.hide();
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

      load();

    },

    open: function (options, callback) {
      this.resourseData = options;
      this.load(options.url, options, callback);
    },

    show: function () {
      this.$bottomsheets.addClass(CLASS_IS_SHOWN).get(0).offsetHeight;
      this.$bottomsheets.addClass(CLASS_IS_SLIDED);
      $('body').addClass(CLASS_OPEN);
    },

    hide: function () {
      var $bottomsheets = this.$bottomsheets;
      var hideEvent;
      var $datePicker = $('.qor-datepicker').not('.hidden');

      if ($datePicker.size()){
        $datePicker.addClass('hidden');
      }

      hideEvent = $.Event(EVENT_HIDE);
      $bottomsheets.trigger(hideEvent);

      if (hideEvent.isDefaultPrevented()) {
        return;
      }

      // empty body html when hide slideout
      this.$body.html('');

      $bottomsheets.
        removeClass(CLASS_IS_SLIDED).
        removeClass(CLASS_IS_SHOWN).
        trigger(EVENT_HIDDEN);

      $('body').removeClass(CLASS_OPEN);

      // reinit bottomsheets template, clear all bind events.
      this.init();

      return false;
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

  QorBottomSheets.DEFAULTS = {
    title: '.qor-form-title, .mdl-layout-title',
    content: false
  };

  QorBottomSheets.TEMPLATE_LOADING = '<div style="text-align: center; margin-top: 30px;"><div class="mdl-spinner mdl-js-spinner is-active qor-layout__bottomsheet-spinner"></div></div>';
  QorBottomSheets.TEMPLATE_SEARCH = (
    '<div class="qor-bottomsheets__search">' +
      '<input autocomplete="off" type="text" class="mdl-textfield__input qor-bottomsheets__search-input" placeholder="Search" />' +
      '<button class="mdl-button mdl-js-button mdl-button--icon qor-bottomsheets__search-button" type="button"><i class="material-icons">search</i></button>' +
    '</div>'
  );



  QorBottomSheets.TEMPLATE = (
    '<div class="qor-bottomsheets">' +
      '<div class="qor-bottomsheets__header">' +
        '<h3 class="qor-bottomsheets__title"></h3>' +
        '<button type="button" class="mdl-button mdl-button--icon mdl-js-button mdl-js-repple-effect qor-bottomsheets__close" data-dismiss="bottomsheets">' +
          '<span class="material-icons">close</span>' +
        '</button>' +
      '</div>' +
      '<div class="qor-bottomsheets__body"></div>' +
    '</div>'
  );

  QorBottomSheets.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorBottomSheets(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $.fn.qorBottomSheets = QorBottomSheets.plugin;

  return QorBottomSheets;

});
