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
