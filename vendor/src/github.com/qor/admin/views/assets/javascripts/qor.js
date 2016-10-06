// init for slideout after show event
$.fn.qorSliderAfterShow = {};

// change Mustache tags from {{}} to [[]]
window.Mustache.tags = ['[[', ']]'];

$(document).ajaxComplete(function( event, xhr, settings ) {
    if (settings.type == "POST" || settings.type == "PUT") {
        if ($.fn.qorSlideoutBeforeHide) {
            $.fn.qorSlideoutBeforeHide = null;
            window.onbeforeunload = null;
        }
    }

});(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('datepicker', ['jquery'], factory);
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
  var document = window.document;
  var $document = $(document);
  var Number = window.Number;
  var NAMESPACE = 'datepicker';

  // Events
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_KEYUP = 'keyup.' + NAMESPACE;
  var EVENT_FOCUS = 'focus.' + NAMESPACE;
  var EVENT_RESIZE = 'resize.' + NAMESPACE;
  var EVENT_SHOW = 'show.' + NAMESPACE;
  var EVENT_HIDE = 'hide.' + NAMESPACE;
  var EVENT_PICK = 'pick.' + NAMESPACE;

  // RegExps
  var REGEXP_FORMAT = /(y|m|d)+/g;
  var REGEXP_DIGITS = /\d+/g;
  var REGEXP_YEAR = /^\d{2,4}$/;

  // Classes
  var CLASS_INLINE = NAMESPACE + '-inline';
  var CLASS_DROPDOWN = NAMESPACE + '-dropdown';
  var CLASS_TOP_LEFT = NAMESPACE + '-top-left';
  var CLASS_TOP_RIGHT = NAMESPACE + '-top-right';
  var CLASS_BOTTOM_LEFT = NAMESPACE + '-bottom-left';
  var CLASS_BOTTOM_RIGHT = NAMESPACE + '-bottom-right';
  var CLASS_PLACEMENTS = [
        CLASS_TOP_LEFT,
        CLASS_TOP_RIGHT,
        CLASS_BOTTOM_LEFT,
        CLASS_BOTTOM_RIGHT
      ].join(' ');
  var CLASS_HIDE = NAMESPACE + '-hide';

  // Maths
  var min = Math.min;

  // Utilities
  var toString = Object.prototype.toString;

  function typeOf(obj) {
    return toString.call(obj).slice(8, -1).toLowerCase();
  }

  function isString(str) {
    return typeof str === 'string';
  }

  function isNumber(num) {
    return typeof num === 'number' && !isNaN(num);
  }

  function isUndefined(obj) {
    return typeof obj === 'undefined';
  }

  function isDate(date) {
    return typeOf(date) === 'date';
  }

  function toArray(obj, offset) {
    var args = [];

    if (Array.from) {
      return Array.from(obj).slice(offset || 0);
    }

    // This is necessary for IE8
    if (isNumber(offset)) {
      args.push(offset);
    }

    return args.slice.apply(obj, args);
  }

  // Custom proxy to avoid jQuery's guid
  function proxy(fn, context) {
    var args = toArray(arguments, 2);

    return function () {
      return fn.apply(context, args.concat(toArray(arguments)));
    };
  }

  function isLeapYear(year) {
    return (year % 4 === 0 && year % 100 !== 0) || year % 400 === 0;
  }

  function getDaysInMonth(year, month) {
    return [31, (isLeapYear(year) ? 29 : 28), 31, 30, 31, 30, 31, 31, 30, 31, 30, 31][month];
  }

  function parseFormat(format) {
    var source = String(format).toLowerCase();
    var parts = source.match(REGEXP_FORMAT);
    var length;
    var i;

    if (!parts || parts.length === 0) {
      throw new Error('Invalid date format.');
    }

    format = {
      source: source,
      parts: parts
    };

    length = parts.length;

    for (i = 0; i < length; i++) {
      switch (parts[i]) {
        case 'dd':
        case 'd':
          format.hasDay = true;
          break;

        case 'mm':
        case 'm':
          format.hasMonth = true;
          break;

        case 'yyyy':
        case 'yy':
          format.hasYear = true;
          break;

        // No default
      }
    }

    return format;
  }

  function Datepicker(element, options) {
    options = $.isPlainObject(options) ? options : {};

    if (options.language) {
      options = $.extend({}, Datepicker.LANGUAGES[options.language], options);
    }

    this.$element = $(element);
    this.options = $.extend({}, Datepicker.DEFAULTS, options);
    this.isBuilt = false;
    this.isShown = false;
    this.isInput = false;
    this.isInline = false;
    this.initialValue = '';
    this.initialDate = null;
    this.startDate = null;
    this.endDate = null;
    this.init();
  }

  Datepicker.prototype = {
    constructor: Datepicker,

    init: function () {
      var options = this.options;
      var $this = this.$element;
      var startDate = options.startDate;
      var endDate = options.endDate;
      var date = options.date;

      this.$trigger = $(options.trigger || $this);
      this.isInput = $this.is('input') || $this.is('textarea');
      this.isInline = options.inline && (options.container || !this.isInput);
      this.format = parseFormat(options.format);
      this.initialValue = this.getValue();
      date = this.parseDate(date || this.initialValue);

      if (startDate) {
        startDate = this.parseDate(startDate);

        if (date.getTime() < startDate.getTime()) {
          date = new Date(startDate);
        }

        this.startDate = startDate;
      }

      if (endDate) {
        endDate = this.parseDate(endDate);

        if (startDate && endDate.getTime() < startDate.getTime()) {
          endDate = new Date(startDate);
        }

        if (date.getTime() > endDate.getTime()) {
          date = new Date(endDate);
        }

        this.endDate = endDate;
      }

      this.date = date;
      this.viewDate = new Date(date);
      this.initialDate = new Date(this.date);

      this.bind();

      if (options.autoshow || this.isInline) {
        this.show();
      }

      if (options.autopick) {
        this.pick();
      }
    },

    build: function () {
      var options = this.options;
      var $this = this.$element;
      var $picker;

      if (this.isBuilt) {
        return;
      }

      this.isBuilt = true;

      this.$picker = $picker = $(options.template);
      this.$week = $picker.find('[data-view="week"]');

      // Years view
      this.$yearsPicker = $picker.find('[data-view="years picker"]');
      this.$yearsPrev = $picker.find('[data-view="years prev"]');
      this.$yearsNext = $picker.find('[data-view="years next"]');
      this.$yearsCurrent = $picker.find('[data-view="years current"]');
      this.$years = $picker.find('[data-view="years"]');

      // Months view
      this.$monthsPicker = $picker.find('[data-view="months picker"]');
      this.$yearPrev = $picker.find('[data-view="year prev"]');
      this.$yearNext = $picker.find('[data-view="year next"]');
      this.$yearCurrent = $picker.find('[data-view="year current"]');
      this.$months = $picker.find('[data-view="months"]');

      // Days view
      this.$daysPicker = $picker.find('[data-view="days picker"]');
      this.$monthPrev = $picker.find('[data-view="month prev"]');
      this.$monthNext = $picker.find('[data-view="month next"]');
      this.$monthCurrent = $picker.find('[data-view="month current"]');
      this.$days = $picker.find('[data-view="days"]');

      if (this.isInline) {
        $(options.container || $this).append($picker.addClass(CLASS_INLINE));
      } else {
        $(document.body).append($picker.addClass(CLASS_DROPDOWN));
        $picker.addClass(CLASS_HIDE);
      }

      this.fillWeek();
    },

    unbuild: function () {
      if (!this.isBuilt) {
        return;
      }

      this.isBuilt = false;
      this.$picker.remove();
    },

    bind: function () {
      var options = this.options;
      var $this = this.$element;

      if ($.isFunction(options.show)) {
        $this.on(EVENT_SHOW, options.show);
      }

      if ($.isFunction(options.hide)) {
        $this.on(EVENT_HIDE, options.hide);
      }

      if ($.isFunction(options.pick)) {
        $this.on(EVENT_PICK, options.pick);
      }

      if (this.isInput) {
        $this.on(EVENT_KEYUP, $.proxy(this.keyup, this));

        if (!options.trigger) {
          $this.on(EVENT_FOCUS, $.proxy(this.show, this));
        }
      }

      this.$trigger.on(EVENT_CLICK, $.proxy(this.show, this));
    },

    unbind: function () {
      var options = this.options;
      var $this = this.$element;

      if ($.isFunction(options.show)) {
        $this.off(EVENT_SHOW, options.show);
      }

      if ($.isFunction(options.hide)) {
        $this.off(EVENT_HIDE, options.hide);
      }

      if ($.isFunction(options.pick)) {
        $this.off(EVENT_PICK, options.pick);
      }

      if (this.isInput) {
        $this.off(EVENT_KEYUP, this.keyup);

        if (!options.trigger) {
          $this.off(EVENT_FOCUS, this.show);
        }
      }

      this.$trigger.off(EVENT_CLICK, this.show);
    },

    showView: function (view) {
      var $yearsPicker = this.$yearsPicker;
      var $monthsPicker = this.$monthsPicker;
      var $daysPicker = this.$daysPicker;
      var format = this.format;

      if (format.hasYear || format.hasMonth || format.hasDay) {
        switch (Number(view)) {
          case 2:
          case 'years':
            $monthsPicker.addClass(CLASS_HIDE);
            $daysPicker.addClass(CLASS_HIDE);

            if (format.hasYear) {
              this.fillYears();
              $yearsPicker.removeClass(CLASS_HIDE);
            } else {
              this.showView(0);
            }

            break;

          case 1:
          case 'months':
            $yearsPicker.addClass(CLASS_HIDE);
            $daysPicker.addClass(CLASS_HIDE);

            if (format.hasMonth) {
              this.fillMonths();
              $monthsPicker.removeClass(CLASS_HIDE);
            } else {
              this.showView(2);
            }

            break;

          // case 0:
          // case 'days':
          default:
            $yearsPicker.addClass(CLASS_HIDE);
            $monthsPicker.addClass(CLASS_HIDE);

            if (format.hasDay) {
              this.fillDays();
              $daysPicker.removeClass(CLASS_HIDE);
            } else {
              this.showView(1);
            }
        }
      }
    },

    hideView: function () {
      if (this.options.autohide) {
        this.hide();
      }
    },

    place: function () {
      var options = this.options;
      var $this = this.$element;
      var $picker = this.$picker;
      var containerWidth = $document.outerWidth();
      var containerHeight = $document.outerHeight();
      var elementWidth = $this.outerWidth();
      var elementHeight = $this.outerHeight();
      var width = $picker.width();
      var height = $picker.height();
      var offsets = $this.offset();
      var left = offsets.left;
      var top = offsets.top;
      var offset = parseFloat(options.offset) || 10;
      var placement = CLASS_TOP_LEFT;

      if (top > height && top + elementHeight + height > containerHeight) {
        top -= height + offset;
        placement = CLASS_BOTTOM_LEFT;
      } else {
        top += elementHeight + offset;
      }

      if (left + width > containerWidth) {
        left = left + elementWidth - width;
        placement = placement.replace('left', 'right');
      }

      $picker.removeClass(CLASS_PLACEMENTS).addClass(placement).css({
        top: top,
        left: left,
        zIndex: parseInt(options.zIndex, 10)
      });
    },

    // A shortcut for triggering custom events
    trigger: function (type, data) {
      var e = $.Event(type, data);

      this.$element.trigger(e);

      return e;
    },

    createItem: function (data) {
      var options = this.options;
      var itemTag = options.itemTag;
      var defaults = {
            text: '',
            view: '',
            muted: false,
            picked: false,
            disabled: false
          };

      $.extend(defaults, data);

      return (
        '<' + itemTag + ' ' +
        (defaults.disabled ? 'class="' + options.disabledClass + '"' :
        defaults.picked ? 'class="' + options.pickedClass + '"' :
        defaults.muted ? 'class="' + options.mutedClass + '"' : '') +
        (defaults.view ? ' data-view="' + defaults.view + '"' : '') +
        '>' +
        defaults.text +
        '</' + itemTag + '>'
      );
    },

    fillAll: function () {
      this.fillYears();
      this.fillMonths();
      this.fillDays();
    },

    fillWeek: function () {
      var options = this.options;
      var weekStart = parseInt(options.weekStart, 10) % 7;
      var days = options.daysMin;
      var list = '';
      var i;

      days = $.merge(days.slice(weekStart), days.slice(0, weekStart));

      for (i = 0; i <= 6; i++) {
        list += this.createItem({
          text: days[i]
        });
      }

      this.$week.html(list);
    },

    fillYears: function () {
      var options = this.options;
      var disabledClass = options.disabledClass || '';
      var suffix = options.yearSuffix || '';
      var filter = $.isFunction(options.filter) && options.filter;
      var startDate = this.startDate;
      var endDate = this.endDate;
      var viewDate = this.viewDate;
      var viewYear = viewDate.getFullYear();
      var viewMonth = viewDate.getMonth();
      var viewDay = viewDate.getDate();
      var date = this.date;
      var year = date.getFullYear();
      var isPrevDisabled = false;
      var isNextDisabled = false;
      var isDisabled = false;
      var isPicked = false;
      var isMuted = false;
      var list = '';
      var start = -5;
      var end = 6;
      var i;

      for (i = start; i <= end; i++) {
        date = new Date(viewYear + i, viewMonth, viewDay);
        isMuted = i === start || i === end;
        isPicked = (viewYear + i) === year;
        isDisabled = false;

        if (startDate) {
          isDisabled = date.getFullYear() < startDate.getFullYear();

          if (i === start) {
            isPrevDisabled = isDisabled;
          }
        }

        if (!isDisabled && endDate) {
          isDisabled = date.getFullYear() > endDate.getFullYear();

          if (i === end) {
            isNextDisabled = isDisabled;
          }
        }

        if (!isDisabled && filter) {
          isDisabled = filter.call(this.$element, date) === false;
        }

        list += this.createItem({
          text: viewYear + i,
          view: isDisabled ? 'year disabled' : isPicked ? 'year picked' : 'year',
          muted: isMuted,
          picked: isPicked,
          disabled: isDisabled
        });
      }

      this.$yearsPrev.toggleClass(disabledClass, isPrevDisabled);
      this.$yearsNext.toggleClass(disabledClass, isNextDisabled);
      this.$yearsCurrent.
        toggleClass(disabledClass, true).
        html((viewYear + start) + suffix + ' - ' + (viewYear + end) + suffix);
      this.$years.html(list);
    },

    fillMonths: function () {
      var options = this.options;
      var disabledClass = options.disabledClass || '';
      var months = options.monthsShort;
      var filter = $.isFunction(options.filter) && options.filter;
      var startDate = this.startDate;
      var endDate = this.endDate;
      var viewDate = this.viewDate;
      var viewYear = viewDate.getFullYear();
      var viewDay = viewDate.getDate();
      var date = this.date;
      var year = date.getFullYear();
      var month = date.getMonth();
      var isPrevDisabled = false;
      var isNextDisabled = false;
      var isDisabled = false;
      var isPicked = false;
      var list = '';
      var i;

      for (i = 0; i <= 11; i++) {
        date = new Date(viewYear, i, viewDay);
        isPicked = viewYear === year && i === month;
        isDisabled = false;

        if (startDate) {
          isPrevDisabled = date.getFullYear() === startDate.getFullYear();
          isDisabled = isPrevDisabled && date.getMonth() < startDate.getMonth();
        }

        if (!isDisabled && endDate) {
          isNextDisabled = date.getFullYear() === endDate.getFullYear();
          isDisabled = isNextDisabled && date.getMonth() > endDate.getMonth();
        }

        if (!isDisabled && filter) {
          isDisabled = filter.call(this.$element, date) === false;
        }

        list += this.createItem({
          index: i,
          text: months[i],
          view: isDisabled ? 'month disabled' : isPicked ? 'month picked' : 'month',
          picked: isPicked,
          disabled: isDisabled
        });
      }

      this.$yearPrev.toggleClass(disabledClass, isPrevDisabled);
      this.$yearNext.toggleClass(disabledClass, isNextDisabled);
      this.$yearCurrent.
        toggleClass(disabledClass, isPrevDisabled && isNextDisabled).
        html(viewYear + options.yearSuffix || '');
      this.$months.html(list);
    },

    fillDays: function () {
      var options = this.options;
      var disabledClass = options.disabledClass || '';
      var suffix = options.yearSuffix || '';
      var months = options.monthsShort;
      var weekStart = parseInt(options.weekStart, 10) % 7;
      var filter = $.isFunction(options.filter) && options.filter;
      var startDate = this.startDate;
      var endDate = this.endDate;
      var viewDate = this.viewDate;
      var viewYear = viewDate.getFullYear();
      var viewMonth = viewDate.getMonth();
      var prevViewYear = viewYear;
      var prevViewMonth = viewMonth;
      var nextViewYear = viewYear;
      var nextViewMonth = viewMonth;
      var date = this.date;
      var year = date.getFullYear();
      var month = date.getMonth();
      var day = date.getDate();
      var isPrevDisabled = false;
      var isNextDisabled = false;
      var isDisabled = false;
      var isPicked = false;
      var prevItems = [];
      var nextItems = [];
      var items = [];
      var total = 42; // 6 rows and 7 columns on the days picker
      var length;
      var i;
      var n;

      // Days of previous month
      // -----------------------------------------------------------------------

      if (viewMonth === 0) {
        prevViewYear -= 1;
        prevViewMonth = 11;
      } else {
        prevViewMonth -= 1;
      }

      // The length of the days of previous month
      length = getDaysInMonth(prevViewYear, prevViewMonth);

      // The first day of current month
      date = new Date(viewYear, viewMonth, 1);

      // The visible length of the days of previous month
      // [0,1,2,3,4,5,6] - [0,1,2,3,4,5,6] => [-6,-5,-4,-3,-2,-1,0,1,2,3,4,5,6]
      n = date.getDay() - weekStart;

      // [-6,-5,-4,-3,-2,-1,0,1,2,3,4,5,6] => [1,2,3,4,5,6,7]
      if (n <= 0) {
        n += 7;
      }

      if (startDate) {
        isPrevDisabled = date.getTime() <= startDate.getTime();
      }

      for (i = length - (n - 1); i <= length; i++) {
        date = new Date(prevViewYear, prevViewMonth, i);
        isDisabled = false;

        if (startDate) {
          isDisabled = date.getTime() < startDate.getTime();
        }

        if (!isDisabled && filter) {
          isDisabled = filter.call(this.$element, date) === false;
        }

        prevItems.push(this.createItem({
          text: i,
          view: 'day prev',
          muted: true,
          disabled: isDisabled
        }));
      }

      // Days of next month
      // -----------------------------------------------------------------------

      if (viewMonth === 11) {
        nextViewYear += 1;
        nextViewMonth = 0;
      } else {
        nextViewMonth += 1;
      }

      // The length of the days of current month
      length = getDaysInMonth(viewYear, viewMonth);

      // The visible length of next month
      n = total - (prevItems.length + length);

      // The last day of current month
      date = new Date(viewYear, viewMonth, length);

      if (endDate) {
        isNextDisabled = date.getTime() >= endDate.getTime();
      }

      for (i = 1; i <= n; i++) {
        date = new Date(nextViewYear, nextViewMonth, i);
        isDisabled = false;

        if (endDate) {
          isDisabled = date.getTime() > endDate.getTime();
        }

        if (!isDisabled && filter) {
          isDisabled = filter.call(this.$element, date) === false;
        }

        nextItems.push(this.createItem({
          text: i,
          view: 'day next',
          muted: true,
          disabled: isDisabled
        }));
      }

      // Days of current month
      // -----------------------------------------------------------------------

      for (i = 1; i <= length; i++) {
        date = new Date(viewYear, viewMonth, i);
        isPicked = viewYear === year && viewMonth === month && i === day;
        isDisabled = false;

        if (startDate) {
          isDisabled = date.getTime() < startDate.getTime();
        }

        if (!isDisabled && endDate) {
          isDisabled = date.getTime() > endDate.getTime();
        }

        if (!isDisabled && filter) {
          isDisabled = filter.call(this.$element, date) === false;
        }

        items.push(this.createItem({
          text: i,
          view: isDisabled ? 'day disabled' : isPicked ? 'day picked' : 'day',
          picked: isPicked,
          disabled: isDisabled
        }));
      }

      // Render days picker
      // -----------------------------------------------------------------------

      this.$monthPrev.toggleClass(disabledClass, isPrevDisabled);
      this.$monthNext.toggleClass(disabledClass, isNextDisabled);
      this.$monthCurrent.
        toggleClass(disabledClass, isPrevDisabled && isNextDisabled).
        html(
          options.yearFirst ?
          viewYear + suffix + ' ' + months[viewMonth] :
          months[viewMonth] + ' ' + viewYear + suffix
        );
      this.$days.html(prevItems.join('') + items.join(' ') + nextItems.join(''));
    },

    click: function (e) {
      var $target = $(e.target);
      var viewDate = this.viewDate;
      var viewYear;
      var viewMonth;
      var viewDay;
      var isYear;
      var year;
      var view;

      e.stopPropagation();
      e.preventDefault();

      if ($target.hasClass('disabled')) {
        return;
      }

      viewYear = viewDate.getFullYear();
      viewMonth = viewDate.getMonth();
      viewDay = viewDate.getDate();
      view = $target.data('view');

      switch (view) {
        case 'years prev':
        case 'years next':
          viewYear = view === 'years prev' ? viewYear - 10 : viewYear + 10;
          year = $target.text();
          isYear = REGEXP_YEAR.test(year);

          if (isYear) {
            viewYear = parseInt(year, 10);
            this.date = new Date(viewYear, viewMonth, min(viewDay, 28));
          }

          this.viewDate = new Date(viewYear, viewMonth, min(viewDay, 28));
          this.fillYears();

          if (isYear) {
            this.showView(1);
            this.pick('year');
          }

          break;

        case 'year prev':
        case 'year next':
          viewYear = view === 'year prev' ? viewYear - 1 : viewYear + 1;
          this.viewDate = new Date(viewYear, viewMonth, min(viewDay, 28));
          this.fillMonths();
          break;

        case 'year current':

          if (this.format.hasYear) {
            this.showView(2);
          }

          break;

        case 'year picked':

          if (this.format.hasMonth) {
            this.showView(1);
          } else {
            this.hideView();
          }

          break;

        case 'year':
          viewYear = parseInt($target.text(), 10);
          this.date = new Date(viewYear, viewMonth, min(viewDay, 28));
          this.viewDate = new Date(viewYear, viewMonth, min(viewDay, 28));

          if (this.format.hasMonth) {
            this.showView(1);
          } else {
            this.hideView();
          }

          this.pick('year');
          break;

        case 'month prev':
        case 'month next':
          viewMonth = view === 'month prev' ? viewMonth - 1 : view === 'month next' ? viewMonth + 1 : viewMonth;
          this.viewDate = new Date(viewYear, viewMonth, min(viewDay, 28));
          this.fillDays();
          break;

        case 'month current':

          if (this.format.hasMonth) {
            this.showView(1);
          }

          break;

        case 'month picked':

          if (this.format.hasDay) {
            this.showView(0);
          } else {
            this.hideView();
          }

          break;

        case 'month':
          viewMonth = $.inArray($target.text(), this.options.monthsShort);
          this.date = new Date(viewYear, viewMonth, min(viewDay, 28));
          this.viewDate = new Date(viewYear, viewMonth, min(viewDay, 28));

          if (this.format.hasDay) {
            this.showView(0);
          } else {
            this.hideView();
          }

          this.pick('month');
          break;

        case 'day prev':
        case 'day next':
        case 'day':
          viewMonth = view === 'day prev' ? viewMonth - 1 : view === 'day next' ? viewMonth + 1 : viewMonth;
          viewDay = parseInt($target.text(), 10);
          this.date = new Date(viewYear, viewMonth, viewDay);
          this.viewDate = new Date(viewYear, viewMonth, viewDay);
          this.fillDays();

          if (view === 'day') {
            this.hideView();
          }

          this.pick('day');
          break;

        case 'day picked':
          this.hideView();
          this.pick('day');
          break;

        // No default
      }
    },

    clickDoc: function (e) {
      var target = e.target;
      var trigger = this.$trigger[0];
      var ignored;

      while (target !== document) {
        if (target === trigger) {
          ignored = true;
          break;
        }

        target = target.parentNode;
      }

      if (!ignored) {
        this.hide();
      }
    },

    keyup: function () {
      this.update();
    },

    getValue: function () {
      var $this = this.$element;
      var val = '';

      if (this.isInput) {
        val = $this.val();
      } else if (this.isInline) {
        if (this.options.container) {
          val = $this.text();
        }
      } else {
        val = $this.text();
      }

      return val;
    },

    setValue: function (val) {
      var $this = this.$element;

      val = isString(val) ? val : '';

      if (this.isInput) {
        $this.val(val);
      } else if (this.isInline) {
        if (this.options.container) {
          $this.text(val);
        }
      } else {
        $this.text(val);
      }
    },


    // Methods
    // -------------------------------------------------------------------------

    // Show the datepicker
    show: function () {
      if (!this.isBuilt) {
        this.build();
      }

      if (this.isShown) {
        return;
      }

      if (this.trigger(EVENT_SHOW).isDefaultPrevented()) {
        return;
      }

      this.isShown = true;
      this.$picker.removeClass(CLASS_HIDE).on(EVENT_CLICK, $.proxy(this.click, this));
      this.showView(this.options.startView);

      if (!this.isInline) {
        $window.on(EVENT_RESIZE, (this._place = proxy(this.place, this)));
        $document.on(EVENT_CLICK, (this._clickDoc = proxy(this.clickDoc, this)));
        this.place();
      }
    },

    // Hide the datepicker
    hide: function () {
      if (!this.isShown) {
        return;
      }

      if (this.trigger(EVENT_HIDE).isDefaultPrevented()) {
        return;
      }

      this.isShown = false;
      this.$picker.addClass(CLASS_HIDE).off(EVENT_CLICK, this.click);

      if (!this.isInline) {
        $window.off(EVENT_RESIZE, this._place);
        $document.off(EVENT_CLICK, this._clickDoc);
      }
    },

    // Update the datepicker with the current input value
    update: function () {
      this.setDate(this.getValue(), true);
    },

    /**
     * Pick the current date to the element
     *
     * @param {String} _view (private)
     */
    pick: function (_view) {
      var $this = this.$element;
      var date = this.date;

      if (this.trigger(EVENT_PICK, {
        view: _view || '',
        date: date
      }).isDefaultPrevented()) {
        return;
      }

      this.setValue(date = this.formatDate(this.date));

      if (this.isInput) {
        $this.trigger('change');
      }
    },

    // Reset the datepicker
    reset: function () {
      this.setDate(this.initialDate, true);
      this.setValue(this.initialValue);

      if (this.isShown) {
        this.showView(this.options.startView);
      }
    },

    /**
     * Get the month name with given argument or the current date
     *
     * @param {Number} month (optional)
     * @param {Boolean} short (optional)
     * @return {String} (month name)
     */
    getMonthName: function (month, short) {
      var options = this.options;
      var months = options.months;

      if ($.isNumeric(month)) {
        month = Number(month);
      } else if (isUndefined(short)) {
        short = month;
      }

      if (short === true) {
        months = options.monthsShort;
      }

      return months[isNumber(month) ? month : this.date.getMonth()];
    },

    /**
     * Get the day name with given argument or the current date
     *
     * @param {Number} day (optional)
     * @param {Boolean} short (optional)
     * @param {Boolean} min (optional)
     * @return {String} (day name)
     */
    getDayName: function (day, short, min) {
      var options = this.options;
      var days = options.days;

      if ($.isNumeric(day)) {
        day = Number(day);
      } else {
        if (isUndefined(min)) {
          min = short;
        }

        if (isUndefined(short)) {
          short = day;
        }
      }

      days = min === true ? options.daysMin : short === true ? options.daysShort : days;

      return days[isNumber(day) ? day : this.date.getDay()];
    },

    /**
     * Get the current date
     *
     * @param {Boolean} formated (optional)
     * @return {Date|String} (date)
     */
    getDate: function (formated) {
      var date = this.date;

      return formated ? this.formatDate(date) : new Date(date);
    },

    /**
     * Set the current date with a new date
     *
     * @param {Date} date
     * @param {Boolean} _isUpdated (private)
     */
    setDate: function (date, _isUpdated) {
      var filter = this.options.filter;

      if (isDate(date) || isString(date)) {
        date = this.parseDate(date);

        if ($.isFunction(filter) && filter.call(this.$element, date) === false) {
          return;
        }

        this.date = date;
        this.viewDate = new Date(date);

        if (!_isUpdated) {
          this.pick();
        }

        if (this.isBuilt) {
          this.fillAll();
        }
      }
    },

    /**
     * Set the start view date with a new date
     *
     * @param {Date} date
     */
    setStartDate: function (date) {
      if (isDate(date) || isString(date)) {
        this.startDate = this.parseDate(date);

        if (this.isBuilt) {
          this.fillAll();
        }
      }
    },

    /**
     * Set the end view date with a new date
     *
     * @param {Date} date
     */
    setEndDate: function (date) {
      if (isDate(date) || isString(date)) {
        this.endDate = this.parseDate(date);

        if (this.isBuilt) {
          this.fillAll();
        }
      }
    },

    /**
     * Parse a date string with the set date format
     *
     * @param {String} date
     * @return {Date} (parsed date)
     */
    parseDate: function (date) {
      var format = this.format;
      var parts = [];
      var length;
      var year;
      var day;
      var month;
      var val;
      var i;

      if (isDate(date)) {
        return new Date(date.getFullYear(), date.getMonth(), date.getDate());
      } else if (isString(date)) {
        parts = date.match(REGEXP_DIGITS) || [];
      }

      date = new Date();
      year = date.getFullYear();
      day = date.getDate();
      month = date.getMonth();
      length = format.parts.length;

      if (parts.length === length) {
        for (i = 0; i < length; i++) {
          val = parseInt(parts[i], 10) || 1;

          switch (format.parts[i]) {
            case 'dd':
            case 'd':
              day = val;
              break;

            case 'mm':
            case 'm':
              month = val - 1;
              break;

            case 'yy':
              year = 2000 + val;
              break;

            case 'yyyy':
              year = val;
              break;

            // No default
          }
        }
      }

      return new Date(year, month, day);
    },

    /**
     * Format a date object to a string with the set date format
     *
     * @param {Date} date
     * @return {String} (formated date)
     */
    formatDate: function (date) {
      var format = this.format;
      var formated = '';
      var length;
      var year;
      var part;
      var val;
      var i;

      if (isDate(date)) {
        formated = format.source;
        year = date.getFullYear();
        val = {
          d: date.getDate(),
          m: date.getMonth() + 1,
          yy: year.toString().substring(2),
          yyyy: year
        };

        val.dd = (val.d < 10 ? '0' : '') + val.d;
        val.mm = (val.m < 10 ? '0' : '') + val.m;
        length = format.parts.length;

        for (i = 0; i < length; i++) {
          part = format.parts[i];
          formated = formated.replace(part, val[part]);
        }
      }

      return formated;
    },

    // Destroy the datepicker and remove the instance from the target element
    destroy: function () {
      this.unbind();
      this.unbuild();
      this.$element.removeData(NAMESPACE);
    }
  };

  Datepicker.LANGUAGES = {};

  Datepicker.DEFAULTS = {
    // Show the datepicker automatically when initialized
    autoshow: false,

    // Hide the datepicker automatically when picked
    autohide: false,

    // Pick the initial date automatically when initialized
    autopick: false,

    // Enable inline mode
    inline: false,

    // A element (or selector) for putting the datepicker
    container: null,

    // A element (or selector) for triggering the datepicker
    trigger: null,

    // The ISO language code (built-in: en-US)
    language: '',

    // The date string format
    format: 'yyyy-mm-dd',

    // The initial date
    date: null,

    // The start view date
    startDate: null,

    // The end view date
    endDate: null,

    // The start view when initialized
    startView: 0, // 0 for days, 1 for months, 2 for years

    // The start day of the week
    weekStart: 0, // 0 for Sunday, 1 for Monday, 2 for Tuesday, 3 for Wednesday, 4 for Thursday, 5 for Friday, 6 for Saturday

    // Show year before month on the datepicker header
    yearFirst: false,

    // A string suffix to the year number.
    yearSuffix: '',

    // Days' name of the week.
    days: ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'],

    // Shorter days' name
    daysShort: ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'],

    // Shortest days' name
    daysMin: ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'],

    // Months' name
    months: ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'],

    // Shorter months' name
    monthsShort: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'],

    // A element tag for each item of years, months and days
    itemTag: 'li',

    // A class (CSS) for muted date item
    mutedClass: 'muted',

    // A class (CSS) for picked date item
    pickedClass: 'picked',

    // A class (CSS) for disabled date item
    disabledClass: 'disabled',

    // The template of the datepicker
    template: (
      '<div class="datepicker-container">' +
        '<div class="datepicker-panel" data-view="years picker">' +
          '<ul>' +
            '<li data-view="years prev">&lsaquo;</li>' +
            '<li data-view="years current"></li>' +
            '<li data-view="years next">&rsaquo;</li>' +
          '</ul>' +
          '<ul data-view="years"></ul>' +
        '</div>' +
        '<div class="datepicker-panel" data-view="months picker">' +
          '<ul>' +
            '<li data-view="year prev">&lsaquo;</li>' +
            '<li data-view="year current"></li>' +
            '<li data-view="year next">&rsaquo;</li>' +
          '</ul>' +
          '<ul data-view="months"></ul>' +
        '</div>' +
        '<div class="datepicker-panel" data-view="days picker">' +
          '<ul>' +
            '<li data-view="month prev">&lsaquo;</li>' +
            '<li data-view="month current"></li>' +
            '<li data-view="month next">&rsaquo;</li>' +
          '</ul>' +
          '<ul data-view="week"></ul>' +
          '<ul data-view="days"></ul>' +
        '</div>' +
      '</div>'
    ),

    // The offset top or bottom of the datepicker from the element
    offset: 10,

    // The `z-index` of the datepicker
    zIndex: 1000,

    // Filter each date item (return `false` to disable a date item)
    filter: null,

    // Event shortcuts
    show: null,
    hide: null,
    pick: null
  };

  Datepicker.setDefaults = function (options) {
    $.extend(Datepicker.DEFAULTS, $.isPlainObject(options) && options);
  };

  // Save the other datepicker
  Datepicker.other = $.fn.datepicker;

  // Register as jQuery plugin
  $.fn.datepicker = function (option) {
    var args = toArray(arguments, 1);
    var result;

    this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var options;
      var fn;

      if (!data) {
        if (/destroy/.test(option)) {
          return;
        }

        options = $.extend({}, $this.data(), $.isPlainObject(option) && option);
        $this.data(NAMESPACE, (data = new Datepicker(this, options)));
      }

      if (isString(option) && $.isFunction(fn = data[option])) {
        result = fn.apply(data, args);
      }
    });

    return isUndefined(result) ? this : result;
  };

  $.fn.datepicker.Constructor = Datepicker;
  $.fn.datepicker.languages = Datepicker.LANGUAGES;
  $.fn.datepicker.setDefaults = Datepicker.setDefaults;

  // No conflict
  $.fn.datepicker.noConflict = function () {
    $.fn.datepicker = Datepicker.other;
    return this;
  };

});
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
  var NAMESPACE = 'qor.action';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_UNDO = 'undo.' + NAMESPACE;
  var ACTION_FORMS = '.qor-action-forms';
  var ACTION_HEADER = '.qor-page__header';
  var ACTION_BODY = '.qor-page__body';
  var ACTION_BUTTON = '.qor-action-button';
  var MDL_BODY = '.mdl-layout__content';
  var ACTION_SELECTORS = '.qor-actions';
  var ACTION_LINK = 'a.qor-action--button';
  var MENU_ACTIONS = '.qor-table__actions a[data-url]';
  var BUTTON_BULKS = '.qor-action-bulk-buttons';
  var QOR_TABLE = '.qor-table-container';
  var QOR_TABLE_BULK = '.qor-table--bulking';
  var QOR_SEARCH = '.qor-search-container';
  var CLASS_IS_UNDO = 'is_undo';
  var QOR_SLIDEOUT = '.qor-slideout';

  var ACTION_FORM_DATA = 'primary_values[]';

  function QorAction(element, options) {
    this.$element = $(element);
    this.$wrap = $(ACTION_FORMS);
    this.options = $.extend({}, QorAction.DEFAULTS, $.isPlainObject(options) && options);
    this.ajaxForm = {};
    this.init();
  }

  QorAction.prototype = {
    constructor: QorAction,

    init: function () {
      this.bind();
      this.initActions();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, $.proxy(this.click, this));
      $(document)
        .on(EVENT_CLICK, '.qor-table--bulking tr', this.click)
        .on(EVENT_CLICK, ACTION_LINK, this.actionLink);
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.click);

      $(document)
        .off(EVENT_CLICK, '.qor-table--bulking tr', this.click)
        .off(EVENT_CLICK, ACTION_LINK, this.actionLink);

    },

    initActions: function () {
      this.tables = $(QOR_TABLE).find('table').size();

      if (!this.tables) {
        $(BUTTON_BULKS).find('button').attr('disabled', true);
        $(ACTION_LINK).attr('disabled', true);
      }
    },

    collectFormData: function () {
      var checkedInputs = $(QOR_TABLE_BULK).find('.mdl-checkbox__input:checked');
      var formData = [];

      if (checkedInputs.size()){
        checkedInputs.each(function () {
          var id = $(this).closest('tr').data('primary-key');
          if (id){
            formData.push({
              name: ACTION_FORM_DATA,
              value: id.toString()
            });
          }
        });
      }
      this.ajaxForm.formData = formData;
      return this.ajaxForm;
    },

    actionLink: function () {
      // if not in index page
      if (!$(QOR_TABLE).find('table').size()) {
        return false;
      }
    },

    actionSubmit: function (e) {
      var $target = $(e.target);
      this.$actionButton = $target;
      this.submit();
      return false;
    },

    click: function (e) {
      var $target = $(e.target);
      this.$actionButton = $target;

      if ($target.data().ajaxForm) {
        this.collectFormData();
        this.ajaxForm.properties = $target.data();
        this.submit();
        return false;
      }


      if ($target.is('.qor-action--bulk')) {
        this.$wrap.removeClass('hidden');
        $(BUTTON_BULKS).find('button').toggleClass('hidden');
        this.appendTableCheckbox();
        $(QOR_TABLE).addClass('qor-table--bulking');
        $(ACTION_HEADER).find(ACTION_SELECTORS).addClass('hidden');
        $(ACTION_HEADER).find(QOR_SEARCH).addClass('hidden');
      }

      if ($target.is('.qor-action--exit-bulk')) {
        this.$wrap.addClass('hidden');
        $(BUTTON_BULKS).find('button').toggleClass('hidden');
        this.removeTableCheckbox();
        $(QOR_TABLE).removeClass('qor-table--bulking');
        $(ACTION_HEADER).find(ACTION_SELECTORS).removeClass('hidden');
        $(ACTION_HEADER).find(QOR_SEARCH).removeClass('hidden');
      }


      if ($(this).is('tr') && !$target.is('a')) {

        var $firstTd = $(this).find('td').first();

        // Manual make checkbox checked or not
        if ($firstTd.find('.mdl-checkbox__input').get(0)) {
          var hasPopoverForm = $('body').hasClass('qor-bottomsheets-open') || $('body').hasClass('qor-slideout-open');
          var $checkbox = $firstTd.find('.mdl-js-checkbox');
          var slideroutActionForm = $('[data-toggle="qor-action-slideout"]').find('form');
          var formValueInput = slideroutActionForm.find('.js-primary-value');
          var primaryValue = $(this).data('primary-key');
          var $alreadyHaveValue = formValueInput.filter('[value="' + primaryValue + '"]');

          $checkbox.toggleClass('is-checked');
          $firstTd.parents('tr').toggleClass('is-selected');

          var isChecked = $checkbox.hasClass('is-checked');

          $firstTd.find('input').prop('checked', isChecked);

          if (slideroutActionForm.size() && hasPopoverForm){

            if (isChecked && !$alreadyHaveValue.size()){
              slideroutActionForm.prepend('<input class="js-primary-value" type="hidden" name="primary_values[]" value="' + primaryValue + '" />');
            }

            if (!isChecked && $alreadyHaveValue.size()){
              $alreadyHaveValue.remove();
            }

          }

          return false;
        }

      }
    },

    renderFlashMessage: function (data) {
      var flashMessageTmpl = QorAction.FLASHMESSAGETMPL;
      Mustache.parse(flashMessageTmpl);
      return Mustache.render(flashMessageTmpl, data);
    },

    submit: function () {
      var _this = this,
          $parent,
          $element = this.$element,
          $actionButton = this.$actionButton,
          ajaxForm = this.ajaxForm || {},
          properties = ajaxForm.properties || $actionButton.data(),
          url = properties.url,
          undoUrl = properties.undoUrl,
          isUndo = $actionButton.hasClass(CLASS_IS_UNDO),
          isInSlideout = $actionButton.closest(QOR_SLIDEOUT).length,
          needDisableButtons = $element && !isInSlideout;

      if (properties.fromIndex && (!ajaxForm.formData || !ajaxForm.formData.length)){
        window.alert(ajaxForm.properties.errorNoItem);
        return;
      }

      if (properties.confirm && properties.ajaxForm && !properties.fromIndex) {
          if (window.confirm(properties.confirm)) {
            properties = $.extend({}, properties, {
              _method: properties.method
            });

            $.post(properties.url, properties, function () {
              window.location.reload();
            });

            return;

          } else {
            return;
          }
      }

      if (properties.confirm && !window.confirm(properties.confirm)) {
        return;
      }

      if (isUndo) {
        url = properties.undoUrl;
      }

      $.ajax(url, {
        method: properties.method,
        data: ajaxForm.formData,
        dataType: properties.datatype,
        beforeSend: function () {
          if (undoUrl) {
            $actionButton.prop('disabled', true);
          } else if (needDisableButtons){
            _this.switchButtons($element, 1);
          }

        },
        success: function (data) {
          // has undo action
          if (undoUrl) {
            $element.triggerHandler(EVENT_UNDO, [$actionButton, isUndo, data]);
            isUndo ? $actionButton.removeClass(CLASS_IS_UNDO) : $actionButton.addClass(CLASS_IS_UNDO);
            $actionButton.prop('disabled', false);
            return;
          }

          if (properties.fromIndex || properties.fromMenu){
            window.location.reload();
            return;
          } else {
            $('.qor-alert').remove();
            needDisableButtons && _this.switchButtons($element);
            isInSlideout ? $parent = $(QOR_SLIDEOUT) : $parent = $(MDL_BODY);
            $parent.find(ACTION_BODY).prepend(_this.renderFlashMessage(data));
          }

        },
        error: function (xhr, textStatus, errorThrown) {
          if (undoUrl) {
            $actionButton.prop('disabled', false);
          } else if (needDisableButtons){
            _this.switchButtons($element);
          }
          window.alert([textStatus, errorThrown].join(': '));
        }
      });
    },

    switchButtons: function ($element, disbale) {
      var needDisbale = disbale ? true : false;
      $element.find(ACTION_BUTTON).prop('disabled', needDisbale);
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    },

    // Helper
    removeTableCheckbox : function () {
      $('.qor-page__body .mdl-data-table__select').each(function (i, e) { $(e).parents('td').remove(); });
      $('.qor-page__body .mdl-data-table__select').each(function (i, e) { $(e).parents('th').remove(); });
      $('.qor-table-container tr.is-selected').removeClass('is-selected');
      $('.qor-page__body table.mdl-data-table--selectable').removeClass('mdl-data-table--selectable');
      $('.qor-page__body tr.is-selected').removeClass('is-selected');
    },

    appendTableCheckbox : function () {
      // Only value change and the table isn't selectable will add checkboxes
      $('.qor-page__body .mdl-data-table__select').each(function (i, e) { $(e).parents('td').remove(); });
      $('.qor-page__body .mdl-data-table__select').each(function (i, e) { $(e).parents('th').remove(); });
      $('.qor-table-container tr.is-selected').removeClass('is-selected');
      $('.qor-page__body table').addClass('mdl-data-table--selectable');

      // init google material
      new window.MaterialDataTable($('.qor-page__body table').get(0));
      $('thead.is-hidden tr th:not(".mdl-data-table__cell--non-numeric")').clone().prependTo($('thead:not(".is-hidden") tr'));

      var $fixedHeadCheckBox = $('thead:not(".is-fixed") .mdl-checkbox__input');
      var isMediaLibrary = $('.qor-table--medialibrary').size();
      var hasPopoverForm = $('body').hasClass('qor-bottomsheets-open') || $('body').hasClass('qor-slideout-open');

      isMediaLibrary && ($fixedHeadCheckBox = $('thead .mdl-checkbox__input'));

      $fixedHeadCheckBox.on('click', function () {

        if (!isMediaLibrary) {
          $('thead.is-fixed tr th').eq(0).find('label').click();
          $(this).closest('label').toggleClass('is-checked');
        }

        var slideroutActionForm = $('[data-toggle="qor-action-slideout"]').find('form');
        var slideroutActionFormPrimaryValues = slideroutActionForm.find('.js-primary-value');

        if (slideroutActionForm.size() && hasPopoverForm){

          if ($(this).is(':checked')) {
            var allPrimaryValues = $('.qor-table--bulking tbody tr');
            allPrimaryValues.each(function () {
              var primaryValue = $(this).data('primary-key');
              if (primaryValue){
                slideroutActionForm.prepend('<input class="js-primary-value" type="hidden" name="primary_values[]" value="' + primaryValue + '" />');
              }
            });
          } else {
            slideroutActionFormPrimaryValues.remove();
          }
        }

      });
    }

  };

  QorAction.FLASHMESSAGETMPL = (
    '<div class="qor-alert qor-action-alert qor-alert--success [[#error]]qor-alert--error[[/error]]" [[#message]]data-dismissible="true"[[/message]] role="alert">' +
      '<button type="button" class="mdl-button mdl-button--icon" data-dismiss="alert">'  +
        '<i class="material-icons">close</i>'  +
      '</button>'  +
      '<span class="qor-alert-message">'  +
        '[[#message]]' +
          '[[message]]' +
        '[[/message]]' +
        '[[#error]]' +
          '[[error]]' +
        '[[/error]]' +
      '</span>'  +
    '</div>'
  );

  QorAction.DEFAULTS = {
  };

  $.fn.qorSliderAfterShow.qorActionInit = function (url, html) {
    var hasAction = $(html).find('[data-toggle="qor-action-slideout"]').size();
    var $actionForm = $('[data-toggle="qor-action-slideout"]').find('form');
    var $checkedItem = $('.qor-page__body .mdl-checkbox__input:checked');

    if (hasAction && $checkedItem.size()){
      // insert checked value into sliderout form
      $checkedItem.each(function (i, e) {
        var id = $(e).parents('tbody tr').data('primary-key');
        if (id){
          $actionForm.prepend('<input class="js-primary-value" type="hidden" name="primary_values[]" value="' + id + '" />');
        }
      });
    }

  };

  QorAction.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new QorAction(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.call(data);
      }
    });
  };

  $(function () {
    var options = {};
    var selector = '[data-toggle="qor.action.bulk"]';

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorAction.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorAction.plugin.call($(selector, e.target), options);
      }).
      on(EVENT_CLICK, MENU_ACTIONS, function (e) {
        (new QorAction()).actionSubmit(e);
        return false;
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorAction;

});
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

  var NAMESPACE = 'qor.autoheight';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_INPUT = 'input';

  function QorAutoheight(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorAutoheight.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorAutoheight.prototype = {
    constructor: QorAutoheight,

    init: function () {
      var $this = this.$element;

      this.overflow = $this.css('overflow');
      this.paddingTop = parseInt($this.css('padding-top'), 10);
      this.paddingBottom = parseInt($this.css('padding-bottom'), 10);
      $this.css('overflow', 'hidden');
      this.resize();
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_INPUT, $.proxy(this.resize, this));
    },

    unbind: function () {
      this.$element.off(EVENT_INPUT, this.resize);
    },

    resize: function () {
      var $this = this.$element;

      if ($this.is(':hidden')) {
        return;
      }

      $this.height('auto').height($this.prop('scrollHeight') - this.paddingTop - this.paddingBottom);
    },

    destroy: function () {
      this.unbind();
      this.$element.css('overflow', this.overflow).removeData(NAMESPACE);
    }
  };

  QorAutoheight.DEFAULTS = {};

  QorAutoheight.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorAutoheight(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = 'textarea.qor-js-autoheight';

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorAutoheight.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorAutoheight.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorAutoheight;

});
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

  var NAMESPACE = 'qor.datepicker';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CHANGE = 'pick.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;

  var CLASS_EMBEDDED = '.qor-datepicker__embedded';
  var CLASS_SAVE = '.qor-datepicker__save';
  var CLASS_PARENT = '.qor-field__datetimepicker';

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

  function QorDatepicker(element, options) {
    this.$element = $(element);
    this.options = $.extend(true, {}, QorDatepicker.DEFAULTS, $.isPlainObject(options) && options);
    this.date = null;
    this.formatDate = null;
    this.built = false;
    this.pickerData = this.$element.data();
    this.init();
  }

  QorDatepicker.prototype = {
    init: function () {
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, $.proxy(this.show, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.show);
    },

    build: function () {
      var $modal;
      var $ele = this.$element;
      var data = this.pickerData;
      var datepickerOptions = {
            date: $ele.val(),
            inline: true
          };
      var parent = $ele.closest(CLASS_PARENT);
      var $targetInput = parent.find(data.targetInput);

      if (this.built) {
        return;
      }

      this.$modal = $modal = $(replaceText(QorDatepicker.TEMPLATE, this.options.text)).appendTo('body');

      if ($targetInput.size()) {
        datepickerOptions.date = $targetInput.val();
      }

      if (data.targetInput && $targetInput.data().startDate) {
        datepickerOptions.startDate = new Date();
      }

      $modal.
        find(CLASS_EMBEDDED).
          on(EVENT_CHANGE, $.proxy(this.change, this)).
          datepicker(datepickerOptions).
          triggerHandler(EVENT_CHANGE);

      $modal.
        find(CLASS_SAVE).
          on(EVENT_CLICK, $.proxy(this.pick, this));

      this.built = true;
    },

    unbuild: function () {
      if (!this.built) {
        return;
      }

      this.$modal.
        find(CLASS_EMBEDDED).
          off(EVENT_CHANGE, this.change).
          datepicker('destroy').
          end().
        find(CLASS_SAVE).
          off(EVENT_CLICK, this.pick).
          end().
        remove();
    },

    change: function (e) {
      var $modal = this.$modal;
      var $target = $(e.target);
      var date;

      this.date = date = $target.datepicker('getDate');
      this.formatDate = $target.datepicker('getDate', true);

      $modal.find('.qor-datepicker__picked-year').text(date.getFullYear());
      $modal.find('.qor-datepicker__picked-date').text([
        $target.datepicker('getDayName', date.getDay(), true) + ',',
        String($target.datepicker('getMonthName', date.getMonth(), true)),
        date.getDate()
      ].join(' '));
    },

    show: function () {
      if (!this.built) {
        this.build();
      }

      this.$modal.qorModal('show');
    },

    pick: function () {
      var $targetInput = this.$element;
      var targetInputClass = this.pickerData.targetInput;
      var newValue = this.formatDate;

      if (targetInputClass) {
        $targetInput = $targetInput.closest(CLASS_PARENT).find(targetInputClass);

        var regDate = /^\d{4}-\d{1,2}-\d{1,2}/;
        var oldValue = $targetInput.val();
        var hasDate = regDate.test(oldValue);

        if (hasDate) {
          newValue = oldValue.replace(regDate, newValue);
        } else {
          newValue = newValue + ' 00:00';
        }

      }

      $targetInput.val(newValue);
      this.$modal.qorModal('hide');
    },

    destroy: function () {
      this.unbind();
      this.unbuild();
      this.$element.removeData(NAMESPACE);
    }
  };

  QorDatepicker.DEFAULTS = {
    text: {
      title: 'Pick a date',
      ok: 'OK',
      cancel: 'Cancel'
    }
  };

  QorDatepicker.TEMPLATE = (
     '<div class="qor-modal fade qor-datepicker" tabindex="-1" role="dialog" aria-hidden="true">' +
      '<div class="mdl-card mdl-shadow--2dp" role="document">' +
        '<div class="mdl-card__title">' +
          '<h2 class="mdl-card__title-text">${title}</h2>' +
        '</div>' +
        '<div class="mdl-card__supporting-text">' +
          '<div class="qor-datepicker__picked">' +
            '<div class="qor-datepicker__picked-year"></div>' +
            '<div class="qor-datepicker__picked-date"></div>' +
          '</div>' +
          '<div class="qor-datepicker__embedded"></div>' +
        '</div>' +
        '<div class="mdl-card__actions">' +
          '<a class="mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect qor-datepicker__save">${ok}</a>' +
          '<a class="mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect" data-dismiss="modal">${cancel}</a>' +
        '</div>' +
      '</div>' +
    '</div>'
  );

  QorDatepicker.plugin = function (option) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var options;
      var fn;

      if (!data) {
        if (!$.fn.datepicker) {
          return;
        }

        if (/destroy/.test(option)) {
          return;
        }

        options = $.extend(true, {}, $this.data(), typeof option === 'object' && option);
        $this.data(NAMESPACE, (data = new QorDatepicker(this, options)));
      }

      if (typeof option === 'string' && $.isFunction(fn = data[option])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.datepicker"]';

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorDatepicker.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorDatepicker.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorDatepicker;

});(function (factory) {
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

  var dirtyForm = function (ele, options) {
    var hasChangedObj = false;

    if (this instanceof jQuery) {
        options = ele;
        ele = this;

    } else if (!(ele instanceof jQuery)) {
        ele = $(ele);
    }

    ele.each(function (item, element) {
        var $ele = $(element);

        if ($ele.is('form')) {
            if ($ele.hasClass('ignore-dirtyform')) {
                return false;
            }
            hasChangedObj = dirtyForm($ele.find('input:not([type="hidden"]):not(".search-field input"):not(".chosen-search input"):not(".ignore-dirtyform"), textarea, select'), options);
            if (hasChangedObj) {
                return false;
            }
        } else if ($ele.is(':checkbox') || $ele.is(':radio')) {

            if (element.checked != element.defaultChecked) {
                hasChangedObj = true;
                return false;
            }

        } else if ($ele.is('input') || $ele.is('textarea')) {

            if (element.value != element.defaultValue) {
                hasChangedObj = true;
                return false;
            }
        } else if ($ele.is('select')) {

            var option;
            var defaultSelectedIndex = 0;
            var numberOfOptions = element.options.length;

            for (var i = 0; i < numberOfOptions; i++) {
                option = element.options[ i ];
                hasChangedObj = (hasChangedObj || (option.selected != option.defaultSelected));
                if (option.defaultSelected) {
                    defaultSelectedIndex = i;
                }
            }

            if (hasChangedObj && !element.multiple) {
                hasChangedObj = (defaultSelectedIndex != element.selectedIndex);
            }

            if (hasChangedObj) {
                return false;
            }
        }

    });

    return hasChangedObj;

    };

    $.fn.extend({
        dirtyForm : dirtyForm
    });

    $(function () {
        $(document).on('submit', 'form', function () {
            window.onbeforeunload = null;
            $.fn.qorSlideoutBeforeHide = null;
        });

        $(document).on('change', 'form', function () {
            if ($(this).dirtyForm()){
                $.fn.qorSlideoutBeforeHide = true;
                window.onbeforeunload = function () {
                    return "You have unsaved changes on this page. If you leave this page, you will lose all unsaved changes.";
                };
            } else {
                $.fn.qorSlideoutBeforeHide = null;
                window.onbeforeunload = null;
            }
        });
    });
});
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

  var location = window.location;
  var NAMESPACE = 'qor.filter';
  var EVENT_FILTER_CHANGE = 'filterChanged.' + NAMESPACE;
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_CHANGE = 'change.' + NAMESPACE;
  var CLASS_IS_ACTIVE = 'is-active';
  var CLASS_BOTTOMSHEETS = '.qor-bottomsheets';

  function QorFilter(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorFilter.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  function encodeSearch(data, detached) {
    var search = location.search;
    var params;

    if ($.isArray(data)) {
      params = decodeSearch(search);

      $.each(data, function (i, param) {
        i = $.inArray(param, params);

        if (i === -1) {
          params.push(param);
        } else if (detached) {
          params.splice(i, 1);
        }
      });

      search = '?' + params.join('&');
    }

    return search;
  }

  function decodeSearch(search) {
    var data = [];

    if (search && search.indexOf('?') > -1) {
      search = search.split('?')[1];

      if (search && search.indexOf('#') > -1) {
        search = search.split('#')[0];
      }

      if (search) {
        // search = search.toLowerCase();
        data = $.map(search.split('&'), function (n) {
          var param = [];
          var value;

          n = n.split('=');
          if (/page/.test(n[0])){
            return;
          }
          value = n[1];
          param.push(n[0]);

          if (value) {
            value = $.trim(decodeURIComponent(value));

            if (value) {
              param.push(value);
            }
          }

          return param.join('=');
        });
      }
    }

    return data;
  }

  QorFilter.prototype = {
    constructor: QorFilter,

    init: function () {
      // this.parse();
      this.bind();
    },

    bind: function () {
      var options = this.options;

      this.$element.
        on(EVENT_CLICK, options.label, $.proxy(this.toggle, this)).
        on(EVENT_CHANGE, options.group, $.proxy(this.toggle, this));
    },

    unbind: function () {
      this.$element.
        off(EVENT_CLICK, this.toggle).
        off(EVENT_CHANGE, this.toggle);
    },

    toggle: function (e) {
      var $target = $(e.currentTarget);
      var data = [];
      var params;
      var param;
      var search;
      var name;
      var value;
      var index;
      var matched;
      var paramName;

      if ($target.is('select')) {
        params = decodeSearch(location.search);
        paramName = name = $target.attr('name');
        value = $target.val();
        param = [name];

        if (value) {
          param.push(value);
        }

        param = param.join('=');

        if (value) {
          data.push(param);
        }

        $target.children().each(function () {
          var $this = $(this);
          var param = [name];
          var value = $.trim($this.prop('value'));

          if (value) {
            param.push(value);
          }

          param = param.join('=');
          index = $.inArray(param, params);

          if (index > -1) {
            matched = param;
            return false;
          }
        });

        if (matched) {
          data.push(matched);
          search = encodeSearch(data, true);
        } else {
          search = encodeSearch(data);
        }
      } else if ($target.is('a')) {
        e.preventDefault();
        paramName = $target.data().paramName;
        data = decodeSearch($target.attr('href'));
        if ($target.hasClass(CLASS_IS_ACTIVE)) {
          search = encodeSearch(data, true); // set `true` to detach
        } else {
          search = encodeSearch(data);
        }
      }

      if (this.$element.closest(CLASS_BOTTOMSHEETS).length) {
        $(CLASS_BOTTOMSHEETS).trigger(EVENT_FILTER_CHANGE, [search, paramName]);
      } else {
        location.search = search;
      }


    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    }
  };

  QorFilter.DEFAULTS = {
    label: false,
    group: false
  };

  QorFilter.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorFilter(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.filter"]';
    var options = {
          label: 'a',
          group: 'select'
        };

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorFilter.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorFilter.plugin.call($(selector, e.target), options);
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorFilter;

});
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
  var _ = window._;
  var NAMESPACE = 'qor.fixer';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_RESIZE = 'resize.' + NAMESPACE;
  var EVENT_SCROLL = 'scroll.' + NAMESPACE;
  var CLASS_IS_HIDDEN = 'is-hidden';
  var CLASS_IS_FIXED = 'is-fixed';
  var CLASS_HEADER = '.qor-page__header';

  function QorFixer(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorFixer.DEFAULTS, $.isPlainObject(options) && options);
    this.$clone = null;
    this.init();
  }

  QorFixer.prototype = {
    constructor: QorFixer,

    init: function () {
      var options = this.options;
      var $this = this.$element;

      // disable fixer if have multiple tables or in search page or in media library list page
      if ($('.qor-page__body .qor-js-table').size() > 1 || $('.qor-global-search--container').size() > 0 || $this.hasClass('qor-table--medialibrary')) {
        return;
      }

      if ($this.is(':hidden') || $this.find('tbody > tr:visible').length <= 1) {
        return;
      }

      this.$thead = $this.find('thead:first');
      this.$tbody = $this.find('tbody:first');
      this.$header = $(options.header);
      this.$subHeader = $(options.subHeader);
      this.$content = $(options.content);
      this.marginBottomPX = parseInt(this.$subHeader.css('marginBottom'));
      this.paddingHeight = options.paddingHeight;
      this.fixedHeaderWidth = [];
      this.isEqualed = false;

      this.resize();
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, $.proxy(this.check, this));

      this.$content.on(EVENT_SCROLL, $.proxy(this.toggle, this));
      $window.on(EVENT_RESIZE, $.proxy(this.resize, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.check);

      this.$content.
        off(EVENT_SCROLL, this.toggle).
        off(EVENT_RESIZE, this.resize);
    },

    build: function () {
      var $this = this.$element;
      var $thead = this.$thead;
      var $clone = this.$clone;
      var self = this;
      var $items = $thead.find('> tr').children();
      var pageBodyTop = this.$content.offset().top + $(CLASS_HEADER).height();

      if (!$clone) {
        this.$clone = $clone = $thead.clone().prependTo($this).css({ top:pageBodyTop });
      }

      $clone.
        addClass([CLASS_IS_FIXED, CLASS_IS_HIDDEN].join(' ')).
        find('> tr').
          children().
            each(function (i) {
              $(this).outerWidth($items.eq(i).outerWidth());
              self.fixedHeaderWidth.push($(this).outerWidth());
            });
    },

    unbuild: function () {
      this.$clone.remove();
    },

    check: function (e) {
      var $target = $(e.target);
      var checked;

      if ($target.is('.qor-js-check-all')) {
        checked = $target.prop('checked');

        $target.
          closest('thead').
          siblings('thead').
            find('.qor-js-check-all').prop('checked', checked).
            closest('.mdl-checkbox').toggleClass('is-checked', checked);
      }
    },

    toggle: function () {
      var self = this;
      var $clone = this.$clone;
      var $thead = this.$thead;
      var scrollTop = this.$content.scrollTop();
      var scrollLeft = this.$content.scrollLeft();
      var offsetTop = this.$subHeader.outerHeight() + this.paddingHeight + this.marginBottomPX;
      var headerHeight = $('.qor-page__header').outerHeight();

      if (!this.isEqualed){
        this.headerWidth = [];
        var $items = $thead.find('> tr').children();
        $items.each(function () {
          self.headerWidth.push($(this).outerWidth());
        });
        var notEqualWidth = _.difference(self.fixedHeaderWidth, self.headerWidth);
        if (notEqualWidth.length){
          $('thead.is-fixed').find('>tr').children().each(function (i) {
            $(this).outerWidth(self.headerWidth[i]);
          });
          this.isEqualed = true;
        }
      }
      if (scrollTop > offsetTop - headerHeight) {
        $clone.css({ 'margin-left': -scrollLeft }).removeClass(CLASS_IS_HIDDEN);
      } else {
        $clone.css({ 'margin-left' : '0' }).addClass(CLASS_IS_HIDDEN);
      }
    },

    resize: function () {
      this.build();
      this.toggle();
    },

    destroy: function () {
      this.unbind();
      this.unbuild();
      this.$element.removeData(NAMESPACE);
    }
  };

  QorFixer.DEFAULTS = {
    header: false,
    content: false
  };

  QorFixer.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new QorFixer(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.call(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-js-table';
    var options = {
          header: '.mdl-layout__header',
          subHeader: '.qor-page__header',
          content: '.mdl-layout__content',
          paddingHeight: 2 // Fix sub header height bug
        };

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorFixer.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorFixer.plugin.call($(selector, e.target), options);
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorFixer;

});
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

  var componentHandler = window.componentHandler;
  var NAMESPACE = 'qor.material';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_UPDATE = 'update.' + NAMESPACE;
  var SELECTOR_COMPONENT = '[class*="mdl-js"]';

  function enable(target) {

    /*jshint undef:false */
    if (componentHandler) {

      // Enable all MDL (Material Design Lite) components within the target element
      if ($(target).is(SELECTOR_COMPONENT)) {
        componentHandler.upgradeElements(target);
      } else {
        componentHandler.upgradeElements($(SELECTOR_COMPONENT, target).toArray());
      }
    }
  }

  function disable(target) {

    /*jshint undef:false */
    if (componentHandler) {

      // Destroy all MDL (Material Design Lite) components within the target element
      if ($(target).is(SELECTOR_COMPONENT)) {
        componentHandler.downgradeElements(target);
      } else {
        componentHandler.downgradeElements($(SELECTOR_COMPONENT, target).toArray());
      }
    }
  }

  $(function () {
    $(document).
      on(EVENT_ENABLE, function (e) {
        enable(e.target);
      }).
      on(EVENT_DISABLE, function (e) {
        disable(e.target);
      }).
      on(EVENT_UPDATE, function (e) {
        disable(e.target);
        enable(e.target);
      });
  });

});
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
  var NAMESPACE = 'qor.modal';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_KEYUP = 'keyup.' + NAMESPACE;
  var EVENT_SHOW = 'show.' + NAMESPACE;
  var EVENT_SHOWN = 'shown.' + NAMESPACE;
  var EVENT_HIDE = 'hide.' + NAMESPACE;
  var EVENT_HIDDEN = 'hidden.' + NAMESPACE;
  var EVENT_TRANSITION_END = 'transitionend';
  var CLASS_OPEN = 'qor-modal-open';
  var CLASS_SHOWN = 'shown';
  var CLASS_FADE = 'fade';
  var CLASS_IN = 'in';
  var ARIA_HIDDEN = 'aria-hidden';

  function QorModal(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorModal.DEFAULTS, $.isPlainObject(options) && options);
    this.transitioning = false;
    this.fadable = false;
    this.init();
  }

  QorModal.prototype = {
    constructor: QorModal,

    init: function () {
      this.fadable = this.$element.hasClass(CLASS_FADE);

      if (this.options.show) {
        this.show();
      } else {
        this.toggle();
      }
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, $.proxy(this.click, this));

      if (this.options.keyboard) {
        $document.on(EVENT_KEYUP, $.proxy(this.keyup, this));
      }
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.click);

      if (this.options.keyboard) {
        $document.off(EVENT_KEYUP, this.keyup);
      }
    },

    click: function (e) {
      var element = this.$element[0];
      var target = e.target;

      if (target === element && this.options.backdrop) {
        this.hide();
        return;
      }

      while (target !== element) {
        if ($(target).data('dismiss') === 'modal') {
          this.hide();
          break;
        }

        target = target.parentNode;
      }
    },

    keyup: function (e) {
      if (e.which === 27) {
        this.hide();
      }
    },

    show: function (noTransition) {
      var $this = this.$element,
          showEvent;

      if (this.transitioning || $this.hasClass(CLASS_IN)) {
        return;
      }

      showEvent = $.Event(EVENT_SHOW);
      $this.trigger(showEvent);

      if (showEvent.isDefaultPrevented()) {
        return;
      }

      $document.find('body').addClass(CLASS_OPEN);

      /*jshint expr:true */
      $this.addClass(CLASS_SHOWN).scrollTop(0).get(0).offsetHeight; // reflow for transition
      this.transitioning = true;

      if (noTransition || !this.fadable) {
        $this.addClass(CLASS_IN);
        this.shown();
        return;
      }

      $this.one(EVENT_TRANSITION_END, $.proxy(this.shown, this));
      $this.addClass(CLASS_IN);
    },

    shown: function () {
      this.transitioning = false;
      this.bind();
      this.$element.attr(ARIA_HIDDEN, false).trigger(EVENT_SHOWN).focus();
    },

    hide: function (noTransition) {
      var $this = this.$element,
          hideEvent;

      if (this.transitioning || !$this.hasClass(CLASS_IN)) {
        return;
      }

      hideEvent = $.Event(EVENT_HIDE);
      $this.trigger(hideEvent);

      if (hideEvent.isDefaultPrevented()) {
        return;
      }

      $document.find('body').removeClass(CLASS_OPEN);
      this.transitioning = true;

      if (noTransition || !this.fadable) {
        $this.removeClass(CLASS_IN);
        this.hidden();
        return;
      }

      $this.one(EVENT_TRANSITION_END, $.proxy(this.hidden, this));
      $this.removeClass(CLASS_IN);
    },

    hidden: function () {
      this.transitioning = false;
      this.unbind();
      this.$element.removeClass(CLASS_SHOWN).attr(ARIA_HIDDEN, true).trigger(EVENT_HIDDEN);
    },

    toggle: function () {
      if (this.$element.hasClass(CLASS_IN)) {
        this.hide();
      } else {
        this.show();
      }
    },

    destroy: function () {
      this.$element.removeData(NAMESPACE);
    }
  };

  QorModal.DEFAULTS = {
    backdrop: true,
    keyboard: true,
    show: true
  };

  QorModal.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorModal(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $.fn.qorModal = QorModal.plugin;

  $(function () {
    var selector = '.qor-modal';

    $(document).
      on(EVENT_CLICK, '[data-toggle="qor.modal"]', function () {
        var $this = $(this);
        var data = $this.data();
        var $target = $(data.target || $this.attr('href'));

        QorModal.plugin.call($target, $target.data(NAMESPACE) ? 'toggle' : data);
      }).
      on(EVENT_DISABLE, function (e) {
        QorModal.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorModal.plugin.call($(selector, e.target));
      });
  });

  return QorModal;

});
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

  var NAMESPACE = 'qor.replicator';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var IS_TEMPLATE = 'is-template';

  function QorReplicator(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorReplicator.DEFAULTS, $.isPlainObject(options) && options);
    this.index = 0;
    this.init();
  }

  QorReplicator.prototype = {
    constructor: QorReplicator,

    init: function () {
      var $this = this.$element;
      var options = this.options;
      var $all = $this.find(options.itemClass);
      var $template;
      this.isMultipleTemplate = $this.data().isMultiple;

      if (!$all.length) {
        return;
      }

      $template = $all.filter(options.newClass);

      if (!$template.length) {
        return;
      }

      // Should destroy all components here
      $template.trigger('disable');

      this.$template = $template;
      this.multipleTemplates = {};
      var $filteredTemplateHtml = $template.filter($this.children(options.childrenClass).children(options.newClass));

      if (this.isMultipleTemplate) {
        this.$template = $filteredTemplateHtml;
        if ($this.children(options.childrenClass).children(options.itemClass).size()){
          this.template = $filteredTemplateHtml.prop('outerHTML');
          this.parse();
        }
      } else {
        this.template = $filteredTemplateHtml.prop('outerHTML');
        $template.data(IS_TEMPLATE, true).hide();
        this.parse();
      }

      // remove hidden empty template, make sure no empty data submit to DB
      $filteredTemplateHtml.remove();

      this.bind();
    },

    parse: function (hasIndex) {
      var i = 0;
      if (!this.template){
        return;
      }

      this.template = this.template.replace(/(\w+)\="(\S*\[\d+\]\S*)"/g, function (attribute, name, value) {
        value = value.replace(/^(\S*)\[(\d+)\]([^\[\]]*)$/, function (input, prefix, index, suffix) {
          if (input === value) {
            if (name === 'name') {
              i = index;
            }

            return (prefix + '[{{index}}]' + suffix);
          }
        });

        return (name + '="' + value + '"');
      });
      if (hasIndex) {
        return;
      }
      this.index = parseFloat(i);
    },

    bind: function () {
      var options = this.options;

      this.$element.
        on(EVENT_CLICK, options.addClass, $.proxy(this.add, this)).
        on(EVENT_CLICK, options.delClass, $.proxy(this.del, this));
    },

    unbind: function () {
      this.$element.
        off(EVENT_CLICK, this.add).
        off(EVENT_CLICK, this.del);
    },

    add: function (e) {
      var options = this.options;
      var self = this;
      var $target = $(e.target).closest(this.options.addClass);
      var templateName = $target.data().template;
      var parents = $target.closest(this.$element);
      var parentsChildren = parents.children(options.childrenClass);
      var $item = this.$template;

      // For multiple fieldset template
      if (this.isMultipleTemplate) {
        this.$template.each (function () {
          self.multipleTemplates[$(this).data().fieldsetName] = $(this);
        });
      }
      var $muptipleTargetTempalte = this.multipleTemplates[templateName];
      if (this.isMultipleTemplate){
        // For multiple template
        if ($target.length) {
          this.template = $muptipleTargetTempalte.prop('outerHTML');
          this.parse(true);
          $item = $(this.template.replace(/\{\{index\}\}/g, ++this.index));
          for (var dataKey in $target.data()) {
            if (dataKey.match(/^sync/)) {
              var k = dataKey.replace(/^sync/, '');
              $item.find('input[name*=\'.' + k + '\']').val($target.data(dataKey));
            }
          }
          if ($target.closest(options.childrenClass).children('fieldset').size()) {
            $target.closest(options.childrenClass).children('fieldset').last().after($item.show());
          } else {
            // If user delete all template
            parentsChildren.prepend($item.show());
          }
        }
      } else {

        if ($target.length) {
          $item = $(this.template.replace(/\{\{index\}\}/g, ++this.index));
          $target.before($item.show());
        }
      }

      if ($item) {
        // Enable all JavaScript components within the fieldset
        $item.trigger('enable');
      }
      e.stopPropagation();
    },

    del: function (e) {
      var options = this.options;
      var $item = $(e.target).closest(options.itemClass);
      var $alert;

      $item.children(':visible').addClass('hidden').hide();
      $alert = $(options.alertTemplate.replace('{{name}}', this.parseName($item)));
      $alert.find(options.undoClass).one(EVENT_CLICK, function () {
        $item.find('> .qor-fieldset__alert').remove();
        $item.children('.hidden').removeClass('hidden').show();

      });

      $item.append($alert);
    },

    parseName: function ($item) {
      var name = $item.find('input[name]').attr('name');

      if (name) {
        return name.replace(/[^\[\]]+$/, '');
      }
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    }
  };

  QorReplicator.DEFAULTS = {
    itemClass: false,
    newClass: false,
    addClass: false,
    delClass: false,
    childrenClass: false,
    alertTemplate: ''
  };

  QorReplicator.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new QorReplicator(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.call(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-fieldset-container';
    var options = {
          itemClass: '.qor-fieldset',
          newClass: '.qor-fieldset--new',
          addClass: '.qor-fieldset__add',
          delClass: '.qor-fieldset__delete',
          childrenClass: '.qor-field__block',
          undoClass: '.qor-fieldset__undo',
          alertTemplate: (
            '<div class="qor-fieldset__alert">' +
              '<input type="hidden" name="{{name}}._destroy" value="1">' +
              '<button class="mdl-button mdl-button--accent mdl-js-button mdl-js-ripple-effect qor-fieldset__undo" type="button">Undo delete</button>' +
            '</div>'
          )
        };

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorReplicator.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorReplicator.plugin.call($(selector, e.target), options);
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorReplicator;

});
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
  var location = window.location;
  var componentHandler = window.componentHandler;
  var history = window.history;
  var NAMESPACE = 'qor.globalSearch';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;

  var SEARCH_RESOURCE = '.qor-global-search--resource';
  var SEARCH_RESULTS = '.qor-global-search--results';
  var QOR_TABLE = '.qor-table';
  var IS_ACTIVE = 'is-active';

  function QorSearchCenter(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSearchCenter.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorSearchCenter.prototype = {
    constructor: QorSearchCenter,

    init: function () {
      this.bind();
      this.initTab();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, $.proxy(this.click, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.check);
    },

    initTab: function () {
      var locationSearch = location.search;
      var resourceName;
      if (/resource_name/.test(locationSearch)){
        resourceName = locationSearch.match(/resource_name=\w+/g).toString().split('=')[1];
        $(SEARCH_RESOURCE).removeClass(IS_ACTIVE);
        $('[data-resource="' + resourceName + '"]').addClass(IS_ACTIVE);
      }
    },

    click : function (e) {
      var $target = $(e.target);
      var data = $target.data();

      if ($target.is(SEARCH_RESOURCE)){
        var oldUrl = location.href.replace(/#/g, '');
        var newUrl;
        var newResourceName = data.resource;
        var hasResource = /resource_name/.test(oldUrl);
        var hasKeyword = /keyword/.test(oldUrl);
        var resourceParam = 'resource_name=' + newResourceName;
        var searchSymbol = hasKeyword ? '&' : '?keyword=&';

        if (newResourceName){
          if (hasResource){
            newUrl = oldUrl.replace(/resource_name=\w+/g, resourceParam);
          } else {
            newUrl = oldUrl + searchSymbol + resourceParam;
          }
        } else {
          newUrl = oldUrl.replace(/&resource_name=\w+/g, '');
        }

        if (history.pushState){
          this.fetchSearch(newUrl, $target);
        } else {
          location.href = newUrl;
        }

      }
    },

    fetchSearch: function (url,$target) {
      var title = document.title;

      $.ajax(url, {
        method: 'GET',
        dataType: 'html',
        beforeSend: function () {
          $('.mdl-spinner').remove();
          $(SEARCH_RESULTS).prepend('<div class="mdl-spinner mdl-js-spinner is-active"></div>').find('.qor-section').hide();
          componentHandler.upgradeElement(document.querySelector('.mdl-spinner'));
        },
        success: function (html) {
          var result = $(html).find(SEARCH_RESULTS).html();
          $(SEARCH_RESOURCE).removeClass(IS_ACTIVE);
          $target.addClass(IS_ACTIVE);
          // change location URL without refresh page
          history.pushState({ Page: url, Title: title }, title, url);
          $('.mdl-spinner').remove();
          $(SEARCH_RESULTS).removeClass('loading').html(result);
          componentHandler.upgradeElements(document.querySelectorAll(QOR_TABLE));
        },
        error: function (xhr, textStatus, errorThrown) {
          $(SEARCH_RESULTS).find('.qor-section').show();
          $('.mdl-spinner').remove();
          window.alert([textStatus, errorThrown].join(': '));
        }
      });
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    }

  };

  QorSearchCenter.DEFAULTS = {
  };

  QorSearchCenter.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new QorSearchCenter(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.call(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.global.search"]';
    var options = {};

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorSearchCenter.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorSearchCenter.plugin.call($(selector, e.target), options);
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorSearchCenter;

});
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

  var $body = $('body');
  var $document = $(document);
  var Mustache = window.Mustache;
  var NAMESPACE = 'qor.selectone';
  var PARENT_NAMESPACE = 'qor.bottomsheets';
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_RELOAD = 'reload.' + PARENT_NAMESPACE;
  var CLASS_CLEAR_SELECT = '.qor-selected-many__remove';
  var CLASS_UNDO_DELETE = '.qor-selected-many__undo';
  var CLASS_DELETED_ITEM = 'qor-selected-many__deleted';
  var CLASS_SELECT_FIELD = '.qor-field__selected-many';
  var CLASS_SELECT_INPUT = '.qor-field__selectmany-input';
  var CLASS_SELECT_ICON = '.qor-select__select-icon';
  var CLASS_SELECT_HINT = '.qor-selectmany__hint';
  var CLASS_PARENT = '.qor-field__selectmany';
  var CLASS_BOTTOMSHEETS = '.qor-bottomsheets';
  var CLASS_SELECTED = 'is_selected';
  var CLASS_MANY = 'qor-bottomsheets__select-many';


  function QorSelectMany(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSelectMany.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorSelectMany.prototype = {
    constructor: QorSelectMany,

    init: function () {
      this.bind();
    },

    bind: function () {
      $document.on(EVENT_CLICK, '[data-selectmany-url]', this.openBottomSheets.bind(this)).
                on(EVENT_RELOAD, '.' + CLASS_MANY, this.reloadData.bind(this));
      
      this.$element
        .on(EVENT_CLICK, CLASS_CLEAR_SELECT, this.clearSelect.bind(this))
        .on(EVENT_CLICK, CLASS_UNDO_DELETE, this.undoDelete.bind(this));

    },

    clearSelect: function (e) {
      var $target = $(e.target),
          $selectFeild = $target.closest(CLASS_PARENT);

      $target.closest('[data-primary-key]').addClass(CLASS_DELETED_ITEM);
      this.updateSelectInputData($selectFeild);

      return false;
    },

    undoDelete: function (e) {
      var $target = $(e.target),
          $selectFeild = $target.closest(CLASS_PARENT);

      $target.closest('[data-primary-key]').removeClass(CLASS_DELETED_ITEM);
      this.updateSelectInputData($selectFeild);

      return false;
    },

    openBottomSheets: function (e) {
      var data = $(e.target).data();

      this.BottomSheets = $body.data('qor.bottomsheets');
      this.bottomsheetsData = data;

      this.$selector = $(data.selectId);
      this.$selectFeild = this.$selector.closest(CLASS_PARENT).find(CLASS_SELECT_FIELD);

      // select many templates
      this.SELECT_MANY_SELECTED_ICON = $('[name="select-many-selected-icon"]').html();
      this.SELECT_MANY_UNSELECTED_ICON = $('[name="select-many-unselected-icon"]').html();
      this.SELECT_MANY_HINT = $('[name="select-many-hint"]').html();
      this.SELECT_MANY_TEMPLATE = $('[name="select-many-template"]').html();

      data.url = data.selectmanyUrl;

      this.BottomSheets.open(data, this.handleSelectMany.bind(this));

    },

    reloadData: function () {
      this.initItems();
    },

    renderSelectMany: function (data) {
      return Mustache.render(this.SELECT_MANY_TEMPLATE, data);
    },

    renderHint: function (data) {
      return Mustache.render(this.SELECT_MANY_HINT, data);
    },

    initItems: function () {
      var $tr = $(CLASS_BOTTOMSHEETS).find('tbody tr'),
          selectedIconTmpl = this.SELECT_MANY_SELECTED_ICON,
          unSelectedIconTmpl = this.SELECT_MANY_UNSELECTED_ICON,
          selectedIDs = [],
          primaryKey,
          $selectedItems = this.$selectFeild.find('[data-primary-key]').not('.' + CLASS_DELETED_ITEM);

      $selectedItems.each(function () {
        selectedIDs.push($(this).data().primaryKey);
      });

      $tr.each(function () {
        var $this = $(this),
            $td = $this.find('td:first');

        primaryKey = $this.data().primaryKey;

        if (selectedIDs.indexOf(primaryKey) !='-1') {
          $this.addClass(CLASS_SELECTED);
          $td.append(selectedIconTmpl);
        } else {
          $td.append(unSelectedIconTmpl);
        }
      });

      this.updateHint(this.getSelectedItemData());
    },

    getSelectedItemData: function() {
      var selecedItems = this.$selectFeild.find('[data-primary-key]').not('.' + CLASS_DELETED_ITEM);
      return {
        selectedNum: selecedItems.length
      };
    },

    updateHint: function (data) {
      var template;

      $.extend(data, this.bottomsheetsData);
      template = this.renderHint(data);

      $(CLASS_SELECT_HINT).remove();
      $(CLASS_BOTTOMSHEETS).find('.qor-page__body').before(template);
    },

    updateSelectInputData: function ($selectFeild) {
      var $selectList = $selectFeild ?  $selectFeild : this.$selectFeild,
          $selectedItems = $selectList.find('[data-primary-key]').not('.' + CLASS_DELETED_ITEM),
          $selector = $selectFeild ? $selectFeild.find(CLASS_SELECT_INPUT) : this.$selector,
          options = $selector.find('option');

      options.prop('selected', false);

      $selectedItems.each(function () {
        options.filter('[value="' + $(this).data().primaryKey + '"]').prop('selected', true);
      });
    },

    changeIcon: function ($ele, template) {
      $ele.find(CLASS_SELECT_ICON).remove();
      $ele.find('td:first').prepend(template);
    },

    removeItem: function (data) {
      var primaryKey = data.primaryKey;

      this.$selectFeild.find('[data-primary-key="' + primaryKey + '"]').remove();
      this.changeIcon(data.$clickElement, this.SELECT_MANY_UNSELECTED_ICON);
    },

    addItem: function (data, isNewData) {
      var template = this.renderSelectMany(data),
          $option,
          $list = this.$selectFeild.find('[data-primary-key="' + data.primaryKey + '"]');

      if ($list.size()) {
        if ($list.hasClass(CLASS_DELETED_ITEM)) {
          $list.removeClass(CLASS_DELETED_ITEM);
          this.updateSelectInputData();
          this.changeIcon(data.$clickElement, this.SELECT_MANY_SELECTED_ICON);
          return;
        } else {
          return;
        }
      }


      this.$selectFeild.append(template);

      if (isNewData) {
        $option = $(Mustache.render(QorSelectMany.SELECT_MANY_OPTION_TEMPLATE, data));
        this.$selector.append($option);
        $option.prop('selected', true);
        this.BottomSheets.hide();
        return;
      }

      this.changeIcon(data.$clickElement, this.SELECT_MANY_SELECTED_ICON);
    },

    handleSelectMany: function () {
      var $bottomsheets = $(CLASS_BOTTOMSHEETS),
          options = {
            formatOnSelect: this.formatSelectResults.bind(this),  // render selected item after click item lists
            formatOnSubmit: this.formatSubmitResults.bind(this)   // render new items after new item form submitted
          };

      $bottomsheets.qorSelectCore(options).addClass(CLASS_MANY);
      this.initItems();
    },

    formatSelectResults: function (data) {
      this.formatResults(data);
    },

    formatSubmitResults: function (data) {
      this.formatResults(data, true);
    },

    formatResults: function (data, isNewData) {
      if (isNewData) {
        this.addItem(data, true);
        return;
      }

      var $element = data.$clickElement,
          isSelected;

      $element.toggleClass(CLASS_SELECTED);
      isSelected = $element.hasClass(CLASS_SELECTED);

      if (isSelected) {
        this.addItem(data);
      } else {
        this.removeItem(data);
      }

      this.updateHint(this.getSelectedItemData());
      this.updateSelectInputData();

    }

  };

  QorSelectMany.SELECT_MANY_OPTION_TEMPLATE = '<option value="[[ primaryKey ]]" >[[ Name ]]</option>';

  QorSelectMany.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorSelectMany(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.selectmany"]';
    $(document).
      on(EVENT_DISABLE, function (e) {
        QorSelectMany.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorSelectMany.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorSelectMany;

});
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

  var $body = $('body');
  var $document = $(document);
  var Mustache = window.Mustache;
  var NAMESPACE = 'qor.selectone';
  var PARENT_NAMESPACE = 'qor.bottomsheets';
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_RELOAD = 'reload.' + PARENT_NAMESPACE;
  var CLASS_CLEAR_SELECT = '.qor-selected__remove';
  var CLASS_CHANGE_SELECT = '.qor-selected__change';
  var CLASS_SELECT_FIELD = '.qor-field__selected';
  var CLASS_SELECT_INPUT = '.qor-field__selectone-input';
  var CLASS_SELECT_TRIGGER = '.qor-field__selectone-trigger';
  var CLASS_PARENT = '.qor-field__selectone';
  var CLASS_BOTTOMSHEETS = '.qor-bottomsheets';
  var CLASS_SELECTED = 'is_selected';
  var CLASS_ONE = 'qor-bottomsheets__select-one';
  

  function QorSelectOne(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSelectOne.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorSelectOne.prototype = {
    constructor: QorSelectOne,

    init: function () {
      this.bind();
    },

    bind: function () {
      $document.on(EVENT_CLICK, '[data-selectone-url]', this.openBottomSheets.bind(this)).
                on(EVENT_RELOAD, '.' + CLASS_ONE, this.reloadData.bind(this));
      
      this.$element.
        on(EVENT_CLICK, CLASS_CLEAR_SELECT, this.clearSelect).
        on(EVENT_CLICK, CLASS_CHANGE_SELECT, this.changeSelect);
    },

    clearSelect: function () {
      var $target = $(this),
          $parent = $target.closest(CLASS_PARENT);

      $parent.find(CLASS_SELECT_FIELD).remove();
      $parent.find(CLASS_SELECT_INPUT)[0].value = '';
      $parent.find(CLASS_SELECT_TRIGGER).show();

      return false;
    },

    changeSelect: function () {
      var $target = $(this),
          $parent = $target.closest(CLASS_PARENT);

      $parent.find(CLASS_SELECT_TRIGGER).trigger('click');

    },

    openBottomSheets: function (e) {
      var data = $(e.target).data();

      this.BottomSheets = $body.data('qor.bottomsheets');
      this.bottomsheetsData = data;
      data.url = data.selectoneUrl;

      this.SELECT_ONE_SELECTED_ICON = $('[name="select-one-selected-icon"]').html();

      this.BottomSheets.open(data, this.handleSelectOne.bind(this));
    },

    initItem: function () {
      var $selectFeild = $(this.bottomsheetsData.selectId).closest(CLASS_PARENT).find(CLASS_SELECT_FIELD),
          selectedID;

      if (!$selectFeild.length) {
        return;
      }

      selectedID = $selectFeild.data().primaryKey;

      if (selectedID) {
        $(CLASS_BOTTOMSHEETS).find('tr[data-primary-key="' + selectedID + '"]').addClass(CLASS_SELECTED).find('td:first').append(this.SELECT_ONE_SELECTED_ICON);
      }
    },

    reloadData: function () {
      this.initItem();
    },

    renderSelectOne: function (data) {
      return Mustache.render($('[name="select-one-selected-template"]').html(), data);
    },

    handleSelectOne: function () {
      var options = {
        formatOnSelect: this.formatSelectResults.bind(this), //render selected item after click item lists
        formatOnSubmit: this.formatSubmitResults.bind(this)  //render new items after new item form submitted
      };

      $(CLASS_BOTTOMSHEETS).qorSelectCore(options).addClass(CLASS_ONE).data(this.bottomsheetsData);
      this.initItem();
    },

    formatSelectResults: function (data) {
      this.formatResults(data);
    },

    formatSubmitResults: function (data) {
      this.formatResults(data, true);
    },

    formatResults: function (data, isNewData) {
      var template,
          bottomsheetsData = this.bottomsheetsData,
          $select = $(bottomsheetsData.selectId),
          $target = $select.closest(CLASS_PARENT),
          $selectFeild = $target.find(CLASS_SELECT_FIELD);

      $select[0].value = data.primaryKey;
      template = this.renderSelectOne(data);

      if ($selectFeild.size()) {
        $selectFeild.remove();
      }

      $target.prepend(template);
      $target.find(CLASS_SELECT_TRIGGER).hide();

      if (isNewData) {
        $select.append(Mustache.render(QorSelectOne.SELECT_ONE_OPTION_TEMPLATE, data));
        $select[0].value = data.primaryKey;
      }

      this.BottomSheets.hide();
    }

  };

  QorSelectOne.SELECT_ONE_OPTION_TEMPLATE = '<option value="[[ primaryKey ]]" >[[ Name ]]</option>';

  QorSelectOne.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorSelectOne(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.selectone"]';
    $(document).
      on(EVENT_DISABLE, function (e) {
        QorSelectOne.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorSelectOne.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorSelectOne;

});
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
  var NAMESPACE = 'qor.selector';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_SELECTOR_CHANGE = 'selectorChanged.' + NAMESPACE;
  var CLASS_OPEN = 'open';
  var CLASS_ACTIVE = 'active';
  var CLASS_HOVER = 'hover';
  var CLASS_SELECTED = 'selected';
  var CLASS_DISABLED = 'disabled';
  var CLASS_CLEARABLE = 'clearable';
  var SELECTOR_SELECTED = '.' + CLASS_SELECTED;
  var SELECTOR_TOGGLE = '.qor-selector-toggle';
  var SELECTOR_LABEL = '.qor-selector-label';
  var SELECTOR_CLEAR = '.qor-selector-clear';
  var SELECTOR_MENU = '.qor-selector-menu';
  var CLASS_BOTTOMSHEETS = '.qor-bottomsheets';

  function QorSelector(element, options) {
    this.options = options;
    this.$element = $(element);
    this.init();
  }

  QorSelector.prototype = {
    constructor: QorSelector,

    init: function () {
      var $this = this.$element;

      this.placeholder = $this.attr('placeholder') || $this.attr('name') || 'Select';
      this.build();
    },

    build: function () {
      var $this = this.$element;
      var $selector = $(QorSelector.TEMPLATE);
      var alignedClass = this.options.aligned + '-aligned';
      var data = {};
      var eleData = $this.data();
      var hover = eleData.hover;
      var paramName = $this.attr('name');

      this.isBottom = eleData.position == 'bottom';

      hover && $selector.addClass(CLASS_HOVER);

      $selector.addClass(alignedClass).find(SELECTOR_MENU).html(function () {
        var list = [];

        $this.children().each(function () {
          var $this = $(this);
          var selected = $this.attr('selected');
          var disabled = $this.attr('disabled');
          var value = $this.attr('value');
          var label = $this.text();
          var classNames = [];

          if (selected) {
            classNames.push(CLASS_SELECTED);
            data.value = value;
            data.label = label;
            data.paramName = paramName;
          }

          if (disabled) {
            classNames.push(CLASS_DISABLED);
          }

          list.push(
            '<li' +
              (classNames.length ? ' class="' + classNames.join(' ') + '"' : '') +
              ' data-value="' + value + '"' +
              ' data-label="' + label + '"' +
              ' data-param-name="' + paramName + '"' +
            '>' +
              label +
            '</li>'
          );
        });

        return list.join('');
      });

      this.$selector = $selector;
      $this.hide().after($selector);
      $selector.find(SELECTOR_TOGGLE).data('paramName', paramName);
      this.pick(data, true);
      this.bind();
    },

    unbuild: function () {
      this.unbind();
      this.$selector.remove();
      this.$element.show();
    },

    bind: function () {
      this.$selector.on(EVENT_CLICK, $.proxy(this.click, this));
      $document.on(EVENT_CLICK, $.proxy(this.close, this));
    },

    unbind: function () {
      this.$selector.off(EVENT_CLICK, this.click);
      $document.off(EVENT_CLICK, this.close);
    },

    click: function (e) {
      var $target = $(e.target);

      e.stopPropagation();

      if ($target.is(SELECTOR_CLEAR)) {
        this.clear();
      } else if ($target.is('li')) {
        if (!$target.hasClass(CLASS_SELECTED) && !$target.hasClass(CLASS_DISABLED)) {
          this.pick($target.data());
        }

        this.close();
      } else if ($target.closest(SELECTOR_TOGGLE).length) {
        this.open();
      }
    },

    pick: function (data, initialized) {
      var $selector = this.$selector;
      var selected = !!data.value;
      var $element = this.$element;

      $selector.
        find(SELECTOR_TOGGLE).
        toggleClass(CLASS_ACTIVE, selected).
        toggleClass(CLASS_CLEARABLE, selected && this.options.clearable).
          find(SELECTOR_LABEL).
          text(data.label || this.placeholder);

      if (!initialized) {
        $selector.
          find(SELECTOR_MENU).
            children('[data-value="' + data.value + '"]').
            addClass(CLASS_SELECTED).
            siblings(SELECTOR_SELECTED).
            removeClass(CLASS_SELECTED);

        $element.val(data.value);


        if ($element.closest(CLASS_BOTTOMSHEETS).length && !$element.closest('[data-toggle="qor.filter"]').length) {
          // If action is in bottom sheet, will trigger filterChanged.qor.selector event, add passed data.value parameter to event.
          $(CLASS_BOTTOMSHEETS).trigger(EVENT_SELECTOR_CHANGE, [data.value, data.paramName]);
        } else {
          $element.trigger('change');
        }
      }
    },

    clear: function () {
      var $element = this.$element;

      this.$selector.
        find(SELECTOR_TOGGLE).
        removeClass(CLASS_ACTIVE).
        removeClass(CLASS_CLEARABLE).
          find(SELECTOR_LABEL).
          text(this.placeholder).
          end().
        end().
        find(SELECTOR_MENU).
          children(SELECTOR_SELECTED).
          removeClass(CLASS_SELECTED);

      $element.val('').trigger('change');
    },

    open: function () {

      // Close other opened dropdowns first
      $document.triggerHandler(EVENT_CLICK);

      // Open the current dropdown
      this.$selector.addClass(CLASS_OPEN);
      if (this.isBottom) {
        this.$selector.addClass('bottom');
      }
    },

    close: function () {
      this.$selector.removeClass(CLASS_OPEN);
      if (this.isBottom) {
        this.$selector.removeClass('bottom');
      }
    },

    destroy: function () {
      this.unbuild();
      this.$element.removeData(NAMESPACE);
    }
  };

  QorSelector.DEFAULTS = {
    aligned: 'left',
    clearable: false
  };

  QorSelector.TEMPLATE = (
    '<div class="qor-selector">' +
      '<a class="qor-selector-toggle">' +
        '<span class="qor-selector-label"></span>' +
        '<i class="material-icons qor-selector-arrow">arrow_drop_down</i>' +
        '<i class="material-icons qor-selector-clear">clear</i>' +
      '</a>' +
      '<ul class="qor-selector-menu"></ul>' +
    '</div>'
  );

  QorSelector.plugin = function (option) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var options;
      var fn;

      if (!data) {
        if (/destroy/.test(option)) {
          return;
        }

        options = $.extend({}, QorSelector.DEFAULTS, $this.data(), typeof option === 'object' && option);
        $this.data(NAMESPACE, (data = new QorSelector(this, options)));
      }

      if (typeof option === 'string' && $.isFunction(fn = data[option])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.selector"]';

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorSelector.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorSelector.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorSelector;

});
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
  var EVENT_KEYUP = 'keyup.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_SUBMIT = 'submit.' + NAMESPACE;
  var EVENT_SHOW = 'show.' + NAMESPACE;
  var EVENT_SLIDEOUT_SUBMIT_COMPLEMENT = 'slideoutSubmitComplete.' + NAMESPACE;
  var EVENT_SLIDEOUT_CLOSED = 'slideoutClosed.' + NAMESPACE;
  var EVENT_SHOWN = 'shown.' + NAMESPACE;
  var EVENT_HIDE = 'hide.' + NAMESPACE;
  var EVENT_HIDDEN = 'hidden.' + NAMESPACE;
  var EVENT_TRANSITIONEND = 'transitionend';
  var CLASS_OPEN = 'qor-slideout-open';
  var CLASS_MINI = 'qor-slideout-mini';
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
    this.slideoutType = false;
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
        on(EVENT_SUBMIT, 'form', $.proxy(this.submit, this))
        .on(EVENT_CLICK, '[data-dismiss="slideout"]', $.proxy(this.hide, this));

      $document.
        on(EVENT_KEYUP, $.proxy(this.keyup, this));
    },

    unbind: function () {
      this.$slideout.
        off(EVENT_SUBMIT, this.submit);

      $document.
        off(EVENT_KEYUP, this.keyup).
        off(EVENT_CLICK, this.hide);
    },

    keyup: function (e) {
      if (e.which === 27) {
        if ($('.qor-bottomsheets').is(':visible') || $('.qor-modal').is(':visible')) {
          return;
        }

        this.hide();
        this.removeSelectedClass();
      }
    },

    loadScript: function (src, url, response) {
      var script = document.createElement('script');
      script.src = src;
      script.onload = function () {

        // exec qorSliderAfterShow after script loaded
        var qorSliderAfterShow = $.fn.qorSliderAfterShow;
        for (var name in qorSliderAfterShow) {
          if (qorSliderAfterShow.hasOwnProperty(name)) {
            qorSliderAfterShow[name].call(this, url, response);
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

    loadExtraResource: function ($body,$response,url,response) {
      var dataBody = $body;
      dataBody  = dataBody.join('');
      dataBody  = dataBody.replace(/<\s*body/gi, '<div');
      dataBody  = dataBody.replace(/<\s*\/body/gi, '</div');
      var bodyClass = $(dataBody).prop('class');
      $('body').addClass(bodyClass);

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

    },

    removeSelectedClass: function () {
      this.$element.find('[data-url]').removeClass(CLASS_IS_SELECTED);
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
            $.fn.qorSlideoutBeforeHide = null;
          },
          success: function (html) {
            var returnUrl = $form.data('returnUrl');
            var refreshUrl = $form.data('refreshUrl');

            $slideout.trigger(EVENT_SLIDEOUT_SUBMIT_COMPLEMENT);

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
                document.cookie = 'qor-flashes=' + btoa(unescape(encodeURIComponent(JSON.stringify(flashStructs)))) + '; path=' + prefix;
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

              this.slideoutType = $content.find('.qor-form-container').data().slideoutType;

              if (!$content.length) {
                return;
              }

              // Get response body tag: http://stackoverflow.com/questions/7001926/cannot-get-body-element-from-ajax-response
              var bodyHtml = response.match(/<\s*body.*>[\s\S]*<\s*\/body\s*>/ig);
              // if no body tag return
              if (bodyHtml) {
                this.loadExtraResource(bodyHtml,$response,url,response);
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

              // callback for after slider loaded HTML
              var qorSliderAfterShow = $.fn.qorSliderAfterShow;

              if (qorSliderAfterShow) {
                for (var name in qorSliderAfterShow) {
                  if (qorSliderAfterShow.hasOwnProperty(name) && $.isFunction(qorSliderAfterShow[name])) {
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

    open: function (options) {
      this.load(options.url, options.data);
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


      if (this.slideoutType == 'mini') {
        $slideout.addClass(CLASS_MINI);
      } else {
        $slideout.removeClass(CLASS_MINI);
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

      if ($.fn.qorSlideoutBeforeHide) {
        if (window.confirm('You have unsaved changes on this slideout. If you close this slideout, you will lose all unsaved changes!')) {
          this.hideSlideout();
        }
      } else {
        this.hideSlideout();
      }

      this.removeSelectedClass();
    },

    hideSlideout: function () {
      var $slideout = this.$slideout;
      var hideEvent;
      var $datePicker = $('.qor-datepicker').not('.hidden');

      // remove onbeforeunload event
      window.onbeforeunload = null;

      $.fn.qorSlideoutBeforeHide = null;

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

      $slideout.trigger(EVENT_SLIDEOUT_CLOSED);
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
    title: '.qor-form-title, .mdl-layout-title',
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

  $.fn.qorSlideout = QorSlideout.plugin;

  return QorSlideout;

});
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

  var location = window.location;
  var NAMESPACE = 'qor.sorter';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var CLASS_IS_SORTABLE = 'is-sortable';

  function QorSorter(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSorter.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorSorter.prototype = {
    constructor: QorSorter,

    init: function () {
      this.$element.addClass(CLASS_IS_SORTABLE);
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, '> thead > tr > th', $.proxy(this.sort, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.sort);
    },

    sort: function (e) {
      var $target = $(e.currentTarget);
      var orderBy = $target.data('orderBy');
      var search = location.search;
      var param = 'order_by=' + orderBy;

      // Stop when it is not sortable
      if (!orderBy) {
        return;
      }

      if (/order_by/.test(search)) {
        search = search.replace(/order_by(=\w+)?/, function () {
          return param;
        });
      } else {
        search += search.indexOf('?') > -1 ? ('&' + param) : param;
      }

      location.search = search;
    },

    destroy: function () {
      this.unbind();
      this.$element.removeClass(CLASS_IS_SORTABLE).removeData(NAMESPACE);
    }
  };

  QorSorter.DEFAULTS = {};

  QorSorter.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorSorter(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-js-table';

    $(document)
      .on(EVENT_DISABLE, function (e) {
        QorSorter.plugin.call($(selector, e.target), 'destroy');
      })
      .on(EVENT_ENABLE, function (e) {
        QorSorter.plugin.call($(selector, e.target));
      })
      .triggerHandler(EVENT_ENABLE);
  });

  return QorSorter;

});
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

  var _ = window._;
  var $body = $('body');
  var NAMESPACE = 'qor.tabbar';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var CLASS_TAB = '.qor-layout__tab-button';
  var CLASS_TAB_CONTENT = '.qor-layout__tab-content';
  var CLASS_TAB_BAR = '.mdl-layout__tab-bar-container';
  var CLASS_TAB_BAR_RIGHT = '.qor-layout__tab-right';
  var CLASS_TAB_BAR_LEFT = '.qor-layout__tab-left';
  var CLASS_ACTIVE = 'is-active';

  function QorTab(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorTab.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorTab.prototype = {
    constructor: QorTab,

    init: function () {
      this.initTab();
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, CLASS_TAB, this.switchTab.bind(this));
      this.$element.on(EVENT_CLICK, CLASS_TAB_BAR_RIGHT, this.scrollTabRight.bind(this));
      this.$element.on(EVENT_CLICK, CLASS_TAB_BAR_LEFT, this.scrollTabLeft.bind(this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, CLASS_TAB, this.switchTab);
      this.$element.off(EVENT_CLICK, CLASS_TAB_BAR_RIGHT, this.scrollTabRight);
      this.$element.off(EVENT_CLICK, CLASS_TAB_BAR_LEFT, this.scrollTabLeft);
    },

    initTab: function () {
      var data = this.$element.data();

      if (!data.scopeActive) {
        $(CLASS_TAB).first().addClass(CLASS_ACTIVE);
        $body.data('tabScopeActive',$(CLASS_TAB).first().data('name'));
      } else {
        $body.data('tabScopeActive',data.scopeActive);
      }

      this.tabWidth = 0;
      this.slideoutWidth = $(CLASS_TAB_CONTENT).outerWidth();

      _.each($(CLASS_TAB), function(ele) {
        this.tabWidth = this.tabWidth + $(ele).outerWidth();
      }.bind(this));

      if (this.tabWidth > this.slideoutWidth) {
        this.$element.find(CLASS_TAB_BAR).append(QorTab.ARROW_RIGHT);
      }

    },

    scrollTabLeft: function (e) {
      e.stopPropagation();

      var $scrollBar = $(CLASS_TAB_BAR),
          scrollLeft = $scrollBar.scrollLeft(),
          jumpDistance = scrollLeft - this.slideoutWidth;

      if (scrollLeft > 0){
        $scrollBar.animate({scrollLeft:jumpDistance}, 400, function () {

          $(CLASS_TAB_BAR_RIGHT).show();
          if ($scrollBar.scrollLeft() == 0) {
            $(CLASS_TAB_BAR_LEFT).hide();
          }

        });
      }
    },

    scrollTabRight: function (e) {
      e.stopPropagation();

      var $scrollBar = $(CLASS_TAB_BAR),
          scrollLeft = $scrollBar.scrollLeft(),
          tabWidth = this.tabWidth,
          slideoutWidth = this.slideoutWidth,
          jumpDistance = scrollLeft + slideoutWidth;

      if (jumpDistance < tabWidth){
        $scrollBar.animate({scrollLeft:jumpDistance}, 400, function () {

          $(CLASS_TAB_BAR_LEFT).show();
          if ($scrollBar.scrollLeft() + slideoutWidth >= tabWidth) {
            $(CLASS_TAB_BAR_RIGHT).hide();
          }

        });

        !$(CLASS_TAB_BAR_LEFT).size() && this.$element.find(CLASS_TAB_BAR).prepend(QorTab.ARROW_LEFT);
      }
    },

    switchTab: function (e) {
      var $target = $(e.target),
          $element = this.$element,
          data = $target.data(),
          tabScopeActive = $body.data().tabScopeActive,
          isInSlideout = $('.qor-slideout').is(':visible');

      if (!isInSlideout) {
        return;
      }

      if ($target.hasClass(CLASS_ACTIVE)){
        return false;
      }

      $element.find(CLASS_TAB).removeClass(CLASS_ACTIVE);
      $target.addClass(CLASS_ACTIVE);

      $.ajax(data.tabUrl, {
          method: 'GET',
          dataType: 'html',
          processData: false,
          contentType: false,
          beforeSend: function () {
            $('.qor-layout__tab-spinner').remove();
            var $spinner = '<div class="mdl-spinner mdl-js-spinner is-active qor-layout__tab-spinner"></div>';
            $(CLASS_TAB_CONTENT).hide().before($spinner);
            window.componentHandler.upgradeElement($('.qor-layout__tab-spinner')[0]);
          },
          success: function (html) {
            $('.qor-layout__tab-spinner').remove();
            $body.data('tabScopeActive',$target.data('name'));
            var $content = $(html).find(CLASS_TAB_CONTENT).html();
            $(CLASS_TAB_CONTENT).show().html($content).trigger('enable');

          },
          error: function () {
            $('.qor-layout__tab-spinner').remove();
            $body.data('tabScopeActive',tabScopeActive);
          }
        });
      return false;
    },

    destroy: function () {
      this.unbind();
      $body.removeData('tabScopeActive');
    }
  };

  QorTab.ARROW_RIGHT = '<a href="javascript://" class="qor-layout__tab-right"></a>';
  QorTab.ARROW_LEFT = '<a href="javascript://" class="qor-layout__tab-left"></a>';

  QorTab.DEFAULTS = {};

  QorTab.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorTab(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.tab"]';

    $(document)
      .on(EVENT_DISABLE, function (e) {
        QorTab.plugin.call($(selector, e.target), 'destroy');
      })
      .on(EVENT_ENABLE, function (e) {
        QorTab.plugin.call($(selector, e.target));
      })
      .triggerHandler(EVENT_ENABLE);
  });

  return QorTab;

});
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

  var NAMESPACE = 'qor.timepicker';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_FOCUS = 'focus.' + NAMESPACE;
  var EVENT_KEYDOWN = 'keydown.' + NAMESPACE;
  var EVENT_BLUR = 'blur.' + NAMESPACE;
  var EVENT_CHANGE_TIME = 'selectTime.' + NAMESPACE;

  var CLASS_PARENT = '.qor-field__datetimepicker';
  var CLASS_TIME_SELECTED = '.ui-timepicker-selected';

  function QorTimepicker(element, options) {
    this.$element = $(element);
    this.options = $.extend(true, {}, QorTimepicker.DEFAULTS, $.isPlainObject(options) && options);
    this.formatDate = null;
    this.pickerData = this.$element.data();
    this.targetInputClass = this.pickerData.targetInput;
    this.parent = this.$element.closest(CLASS_PARENT);
    this.isDateTimePicker = this.targetInputClass && this.parent.size();
    this.$targetInput = this.parent.find(this.targetInputClass);
    this.init();
  }

  QorTimepicker.prototype = {
    init: function () {
      this.bind();
      this.oldValue = this.$targetInput.val();

      var dateNow = new Date();
      var month = dateNow.getMonth();
      month = (month < 8) ? '0' + (month + 1) : month;

      this.dateValueNow = dateNow.getFullYear() + '-' + month + '-' + dateNow.getDate();

    },

    bind: function () {

      var pickerOptions = {
            timeFormat: 'H:i',
            showOn: null,
            wrapHours: false,
            scrollDefault: 'now'
          };

      if (this.isDateTimePicker) {
        this.$targetInput
          .timepicker(pickerOptions)
          .on(EVENT_CHANGE_TIME, $.proxy(this.changeTime, this))
          .on(EVENT_BLUR, $.proxy(this.blur, this))
          .on(EVENT_FOCUS, $.proxy(this.focus, this))
          .on(EVENT_KEYDOWN, $.proxy(this.keydown, this));
      }

      this.$element.on(EVENT_CLICK, $.proxy(this.show, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.show);

      if (this.isDateTimePicker) {
        this.$targetInput
        .off(EVENT_CHANGE_TIME, this.changeTime)
        .off(EVENT_BLUR, this.blur)
        .off(EVENT_FOCUS, this.focus)
        .off(EVENT_KEYDOWN, this.keydown);
      }
    },

    focus: function () {

    },

    blur: function () {
      var inputValue = this.$targetInput.val();
      var inputArr = inputValue.split(' ');
      var inputArrLen = inputArr.length;

      var tempValue;
      var newDateValue;
      var newTimeValue;
      var isDate;
      var isTime;
      var splitSym;

      var timeReg = /\d{1,2}:\d{1,2}/;
      var dateReg = /^\d{4}-\d{1,2}-\d{1,2}/;

      if (inputArrLen == 1) {
        if (dateReg.test(inputArr[0])) {
          newDateValue = inputArr[0];
          newTimeValue = '00:00';
        }

        if (timeReg.test(inputArr[0])) {
          newDateValue = this.dateValueNow;
          newTimeValue = inputArr[0];
        }

      } else {
        for (var i = 0; i < inputArrLen; i++) {
          // check for date && time
          isDate = dateReg.test(inputArr[i]);
          isTime = timeReg.test(inputArr[i]);

          if (isDate) {
            newDateValue = inputArr[i];
            splitSym = '-';
          }

          if (isTime){
            newTimeValue = inputArr[i];
            splitSym = ':';
          }

          tempValue = inputArr[i].split(splitSym);

          for (var j = 0; j < tempValue.length; j++) {
            if (tempValue[j].length < 2) {
              tempValue[j] = '0' + tempValue[j];
            }
          }

          if (isDate) {
            newDateValue = tempValue.join(splitSym);
          }

          if (isTime) {
            newTimeValue = tempValue.join(splitSym);
          }
        }

      }

      if (this.checkDate(newDateValue) && this.checkTime(newTimeValue)) {
        this.$targetInput.val(newDateValue + ' ' + newTimeValue);
        this.oldValue = this.$targetInput.val();
      } else {
        this.$targetInput.val(this.oldValue);
      }

    },

    keydown: function (e) {
      var keycode = e.keyCode;
      var keys = [48,49,50,51,52,53,54,55,56,57,8,37,38,39,40,27,32,20,189,16,186,96,97,98,99,100,101,102,103,104,105];
      if (keys.indexOf(keycode) == -1) {
        e.preventDefault();
      }
    },

    checkDate: function (value) {
      var regCheckDate = /^(?:(?!0000)[0-9]{4}-(?:(?:0[1-9]|1[0-2])-(?:0[1-9]|1[0-9]|2[0-8])|(?:0[13-9]|1[0-2])-(?:29|30)|(?:0[13578]|1[02])-31)|(?:[0-9]{1,2}(?:0[48]|[2468][048]|[13579][26])|(?:0[48]|[2468][048]|[13579][26])00)-02-29)$/;
      return regCheckDate.test(value);
    },

    checkTime: function (value) {
      var regCheckTime = /^([01]\d|2[0-3]):?([0-5]\d)$/;
      return regCheckTime.test(value);
    },

    changeTime: function () {
      var $targetInput = this.$targetInput;

      var oldValue = this.oldValue;
      var timeReg = /\d{1,2}:\d{1,2}/;
      var hasTime = timeReg.test(oldValue);
      var selectedTime = $targetInput.data().timepickerList.find(CLASS_TIME_SELECTED).html();
      var newValue;

      if (!oldValue) {
        newValue = this.dateValueNow + ' ' + selectedTime;
      } else if (hasTime) {
        newValue = oldValue.replace(timeReg,selectedTime);
      } else {
        newValue = oldValue + ' ' + selectedTime;
      }

      $targetInput.val(newValue);

    },

    show: function () {
      if (!this.isDateTimePicker) {
        return;
      }

      this.$targetInput.timepicker('show');
      this.oldValue = this.$targetInput.val();

    },

    destroy: function () {
      this.unbind();
      this.$targetInput.timepicker('remove');
      this.$element.removeData(NAMESPACE);
    }
  };

  QorTimepicker.DEFAULTS = {};

  QorTimepicker.plugin = function (option) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var options;
      var fn;

      if (!data) {
        if (!$.fn.datepicker) {
          return;
        }

        if (/destroy/.test(option)) {
          return;
        }

        options = $.extend(true, {}, $this.data(), typeof option === 'object' && option);
        $this.data(NAMESPACE, (data = new QorTimepicker(this, options)));
      }

      if (typeof option === 'string' && $.isFunction(fn = data[option])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.timepicker"]';

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorTimepicker.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorTimepicker.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorTimepicker;

});