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
  var NAMESPACE = 'qor.selectcore';
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_SUBMIT = 'submit.' + NAMESPACE;
  var CLASS_TABLE_CONTENT = '.qor-table__content';
  var CLASS_CLICK_TABLE = '.qor-table-container tbody tr';

  function QorSelectCore(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSelectCore.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorSelectCore.prototype = {
    constructor: QorSelectCore,

    init: function () {
      this.bind();
    },

    bind: function () {
      this.$element.
        on(EVENT_CLICK, CLASS_CLICK_TABLE, this.processingData.bind(this)).
        on(EVENT_SUBMIT, 'form', this.submit.bind(this));
    },

    unbind: function () {
      this.$element.
        off(EVENT_CLICK, '.qor-table tbody tr', this.processingData.bind(this)).
        off(EVENT_SUBMIT, 'form', this.submit.bind(this));
    },

    processingData: function (e) {
      var $this = $(e.target).closest('tr'),
          $headings = $this.find('[data-heading]'),
          $heading,
          $content,
          data = {},
          name,
          value,
          options = this.options,
          formatOnSelect = options.formatOnSelect;

      $.extend(data, $this.data());
      data.$clickElement = $this;

      $headings.each(function () {
        $heading = $(this),
        $content = $heading.find(CLASS_TABLE_CONTENT);
        name = $heading.data('heading');
        value = $content.size() ? $content.html() : $heading.html();
        if (name) {
          data[name] = $.trim(value);
        }
      });

      if (formatOnSelect && $.isFunction(formatOnSelect)) {
        formatOnSelect(data);
      }

      return false;
    },

    submit: function (e) {
      var form = e.target;
      var $form = $(form);
      var _this = this;
      var $submit = $form.find(':submit');
      var data;
      var formatOnSubmit = this.options.formatOnSubmit;

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
          success: function (json) {
            data = json;
            data.primaryKey = data.ID;

            if (formatOnSubmit && $.isFunction(formatOnSubmit)) {
              formatOnSubmit(data);
            } else {
              _this.refresh();
            }

          },
          error: function (xhr, textStatus, errorThrown) {

            var $error;
            // Custom HTTP status code
            if (xhr.status === 422) {

              // Clear old errors
              $form.find('.qor-error').remove();
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

    refresh: function () {
      setTimeout(function () {
        window.location.reload();
      }, 350);
    },

    destroy: function () {
      this.unbind();
    }

  };

  QorSelectCore.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }
        $this.data(NAMESPACE, (data = new QorSelectCore(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $.fn.qorSelectCore = QorSelectCore.plugin;

  return QorSelectCore;

});
