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

  var NAMESPACE = 'qor.chooser.sortable';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var CLASS_CHOSE = '.select2-selection__choice';
  var CLASS_CHOSE_REMOVE = '.select2-selection__choice__remove';
  var CLASS_CHOSE_CONTAINER = '.select2-container';
  var CLASS_CHOSE_INPUT = '.select2-search__field';
  var CLASS_SORTABLE_BODY = '.qor-dragable';
  var CLASS_SORTABLE = '.qor-dragable__list';
  var CLASS_SORTABLE_HANDLE = '.qor-dragable__list-handle';
  var CLASS_SORTABLE_DELETE = '.qor-dragable__list-delete';
  var CLASS_SORTABLE_DATA = '.qor-dragable__list-data';
  var CLASS_SORTABLE_BUTTON_ADD = '.qor-dragable__button-add';
  var IS_LOADED = 'sortable-select-many-loaded';

  function QorChooserSortable(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorChooserSortable.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorChooserSortable.prototype = {
    constructor: QorChooserSortable,

    init: function () {
      var $this = this.$element;
      var $parent = $this.parents(CLASS_SORTABLE_BODY);
      var placeholderText = $this.data('placeholder');
      var self = this;

      this.$selector = $parent.find(CLASS_SORTABLE_DATA);
      this.$sortableList = $parent.find(CLASS_SORTABLE);
      this.$parent = $parent;

      var sortEle = $parent.find(CLASS_SORTABLE)[0];

      this.sortable = window.Sortable.create(sortEle, {
          animation: 150,
          handle: CLASS_SORTABLE_HANDLE,
          filter: CLASS_SORTABLE_DELETE,
          dataIdAttr: 'data-index',

          onFilter: function (e){
            var $ele = $(e.item);
            var eleIndex = $ele.data('index');

            $ele.remove();
            self.removeItemsFromList(eleIndex);
          },
          onUpdate: function (){
            self.renderOption();
          }
      });

      $this.select2({
        minimumResultsForSearch: 1,
        dropdownParent: $this.parent()
      })
      .on('change', function () {
        $parent.find(CLASS_CHOSE).hide();
        $(CLASS_CHOSE_INPUT).attr('placeholder',placeholderText);
      })
      .on('select2:select', function (e) {
        self.addItems(e.params.data.id);
      })
      .on('select2:unselect', function (e) {
        self.removeItems(e.params.data.id);
      });

      $parent.find(CLASS_CHOSE_CONTAINER).hide();
      $parent.find(CLASS_CHOSE).hide();
      $(CLASS_CHOSE_INPUT).attr('placeholder',placeholderText);

      this.bind();

    },

    bind: function () {
      this.$parent.on(EVENT_CLICK, CLASS_SORTABLE_BUTTON_ADD,  this.show.bind(this));
    },

    unbind: function () {
      this.$parent.off(EVENT_CLICK, CLASS_SORTABLE_BUTTON_ADD, this.show);
    },

    show: function () {
      var $container = this.$parent.find(CLASS_CHOSE_CONTAINER);

      $container.show();
      this.$parent.find(CLASS_SORTABLE_BUTTON_ADD).hide();
      setTimeout(function(){
        $container.find(CLASS_CHOSE_INPUT).click();
      },100);
    },

    renderItem: function (data) {
      return window.Mustache.render(QorChooserSortable.LIST_HTML, data);
    },

    renderOption: function () {
      var indexArr = this.sortable.toArray();
      var $selector = this.$selector;

      $selector.empty();

      window._.each(indexArr,function(id){
        $selector.append( window.Mustache.render(QorChooserSortable.OPTION_HTML, ({'value': id})));
      });

    },

    removeItems: function (index) {
      $(CLASS_SORTABLE).find('li[data-index="' + index + '"]').remove();
      this.renderOption();
    },

    removeItemsFromList: function (index) {
      this.$parent.find(CLASS_CHOSE).filter('[option-id="'+ index + '"]').find(CLASS_CHOSE_REMOVE).click();
      this.renderOption();
    },

    addItems: function (index) {
      var $this = this.$element;
      var data = $this.find('option[value="' + index + '"]').data();

      this.$sortableList.append(this.renderItem(data));
      this.renderOption();
    },

    destroy: function () {
      this.sortable.destroy();
      this.unbind();
      this.$element.select2('destroy').removeData(NAMESPACE);
    }
  };

  QorChooserSortable.DEFAULTS = {};

  QorChooserSortable.LIST_HTML = '<li data-index=[[index]] data-value=[[value]]><span>[[value]]</span><i class="material-icons qor-dragable__list-delete">clear</i><i class="material-icons qor-dragable__list-handle">drag_handle</i></li>';
  QorChooserSortable.OPTION_HTML = '<option selected value=[[value]]></option>';

  QorChooserSortable.plugin = function (options) {
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

        $this.data(NAMESPACE, (data = new QorChooserSortable(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = 'select[data-toggle="qor.chooser.sortable"]';

    if ($('body').data(IS_LOADED)) {
      return;
    }

    $('body').data(IS_LOADED,true);

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorChooserSortable.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorChooserSortable.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorChooserSortable;

});
