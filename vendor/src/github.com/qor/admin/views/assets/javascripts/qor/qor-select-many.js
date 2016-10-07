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
