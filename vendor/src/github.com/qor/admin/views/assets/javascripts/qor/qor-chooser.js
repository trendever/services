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

  var Mustache = window.Mustache;
  var NAMESPACE = 'qor.chooser';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;

  function QorChooser(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorChooser.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorChooser.prototype = {
    constructor: QorChooser,

    init: function () {
      var $this = this.$element;
      var remoteUrl = $this.data('remote-data-url');
      var option = {
        minimumResultsForSearch: 20,
        dropdownParent: $this.parent()
      };

      if (remoteUrl) {
        option.ajax = {
          url: remoteUrl,
          dataType: 'json',
          cache: true,
          delay: 250,
          data: function (params) {
            return {
              keyword: params.term, // search term
              page: params.page,
              per_page: 20
            };
          },
          processResults: function (data, params) {
            // parse the results into the format expected by Select2
            // since we are using custom formatting functions we do not need to
            // alter the remote JSON data, except to indicate that infinite
            // scrolling can be used
            params.page = params.page || 1;

            var processedData = $.map(data, function (obj) {
              obj.id = obj.Id || obj.ID;
              return obj;
            });

            return {
              results: processedData,
              pagination: {
                more: processedData.length >= 20
              }
            };
          }
        };

        option.templateResult =  function(data) {
          var tmpl = $this.parents('.qor-field').find('[name="select2-result-template"]');
          return QorChooser.formatResult(data, tmpl);
        };

        option.templateSelection = function(data) {
          if (data.loading) return data.text;
          var tmpl = $this.parents('.qor-field').find('[name="select2-selection-template"]');
          return QorChooser.formatResult(data, tmpl);
        };
      }

      $this.on('select2:select', function (evt) {
        $(evt.target).attr('chooser-selected','true');
      }).on('select2:unselect', function (evt) {
        $(evt.target).attr('chooser-selected','');
      });

      $this.select2(option);
    },

    destroy: function () {
      this.$element.select2('destroy').removeData(NAMESPACE);
    }
  };

  QorChooser.DEFAULTS = {};

  QorChooser.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (!$.fn.chosen) {
          return;
        }

        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorChooser(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  QorChooser.formatResult = function (data, tmpl) {
    var result = "";
    if (tmpl.length > 0) {
      result = Mustache.render(tmpl.html().replace(/{{(.*?)}}/g, '[[$1]]'), data);
    } else {
      result = data.text || data.Name || data.Title || data.Code || data[Object.keys(data)[0]];
    }

    // if is HTML
    if (/<(.*)(\/>|<\/.+>)/.test(result)) {
      return $(result);
    }
    return result;
  };

  $(function () {
    var selector = 'select[data-toggle="qor.chooser"]';

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorChooser.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorChooser.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorChooser;

});
