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

  var NAMESPACE = 'qor.chooser';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var CLASS_MULTI = '.chosen-container-multi';
  var CLASS_DEFAULT = 'chosen-default';
  var CLASS_CHOSE = '.search-choice';

  function QorChooser(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorChooser.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorChooser.prototype = {
    constructor: QorChooser,

    init: function () {
      var $this = this.$element;

      if (!$this.prop('multiple')) {
        if ($this.children('[selected]').length) {
          $this.prepend('<option value=""></option>');
        } else {
          $this.prepend('<option value="" selected></option>');
        }
      }

      $this.chosen({
        // jscs:disable requireCamelCaseOrUpperCaseIdentifiers
        allow_single_deselect: true,
        search_contains: true,
        disable_search_threshold: 10,
        width: '100%',
        display_selected_options: false
      })
      .on('change', function (e,params) {
        var $target = $(e.target);
        var $chosenMulti = $target.next(CLASS_MULTI);

        if (!$chosenMulti.size()){
          return;
        }

        if (params.deselected){
          setTimeout(function () {
            if (!$chosenMulti.find(CLASS_CHOSE).size()){
              $chosenMulti.addClass(CLASS_DEFAULT);
            }
          }, 10);
        } else if (params.selected){
          $chosenMulti.removeClass(CLASS_DEFAULT);
        }

      });

      // init multiple selector layout
      if ($this.prop('multiple')){
        var $thisChosenMulti = $this.next(CLASS_MULTI);
        if (!$thisChosenMulti.find(CLASS_CHOSE).size()){
          $thisChosenMulti.addClass(CLASS_DEFAULT);
        }
      }

    },

    destroy: function () {
      this.$element.chosen('destroy').removeData(NAMESPACE);
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
