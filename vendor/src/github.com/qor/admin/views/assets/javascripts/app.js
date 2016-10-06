$(function () {

  'use strict';

  $(document).on('click.qor.alert', '[data-dismiss="alert"]', function () {
    $(this).closest('.qor-alert').remove();
  });

  setTimeout(function () {
    $('.qor-alert[data-dismissible="true"]').remove();
  }, 5000);

});
$(function () {

  'use strict';

  var $form = $('.qor-page__body > .qor-form-container > form');

  $('.qor-error > li > label').each(function () {
    var $label = $(this);
    var id = $label.attr('for');

    if (id) {
      $form.find('#' + id).
        closest('.qor-field').
        addClass('is-error').
        append($label.clone().addClass('qor-field__error'));
    }
  });

});
$(function () {

  'use strict';

  var modal = (
    '<div class="qor-dialog qor-dialog--global-search" tabindex="-1" role="dialog" aria-hidden="true">' +
      '<div class="qor-dialog-content">' +
        '<form action=[[actionUrl]]>' +
          '<div class="mdl-textfield mdl-js-textfield" id="global-search-textfield">' +
            '<input class="mdl-textfield__input ignore-dirtyform" name="keyword" id="globalSearch" value="" type="text" placeholder="" />' +
            '<label class="mdl-textfield__label" for="globalSearch">[[placeholder]]</label>' +
          '</div>' +
        '</form>' +
      '</div>' +
    '</div>'
  );

  $(document).on('click', '.qor-dialog--global-search', function(e){
    e.stopPropagation();
    if (!$(e.target).parents('.qor-dialog-content').size() && !$(e.target).is('.qor-dialog-content')){
      $('.qor-dialog--global-search').remove();
    }
  });

  $(document).on('click', '.qor-global-search--show', function(e){
      e.preventDefault();

      var data = $(this).data();
      var modalHTML = window.Mustache.render(modal, data);

      $('body').append(modalHTML);
      window.componentHandler.upgradeElement(document.getElementById('global-search-textfield'));
      $('#globalSearch').focus();

  });
});
$(function () {

  'use strict';

  $('.qor-menu-container').on('click', '> ul > li > a', function () {
    var $this = $(this);
    var $li = $this.parent();
    var $ul = $this.next('ul');

    if (!$ul.length) {
      return;
    }

    if ($ul.hasClass('in')) {
      $li.removeClass('is-expanded');
      $ul.one('transitionend', function () {
        $ul.removeClass('collapsing in');
      }).addClass('collapsing').height(0);
    } else {
      $li.addClass('is-expanded');
      $ul.one('transitionend', function () {
        $ul.removeClass('collapsing');
      }).addClass('collapsing in').height($ul.prop('scrollHeight'));
    }
  }).find('> ul > li > a').each(function () {
    var $this = $(this);
    var $li = $this.parent();
    var $ul = $this.next('ul');

    if (!$ul.length) {
      return;
    }

    $li.addClass('has-menu is-expanded');
    $ul.addClass('collapse in').height($ul.prop('scrollHeight'));
  });

  if ($('.qor-page').find('.qor-page__header').size()){
    $('.qor-page').addClass("has-header");
    $('header.mdl-layout__header').addClass('has-action');
  }

});
$(function () {
  $('.qor-mobile--show-actions').on('click', function () {
    $('.qor-page__header').toggleClass('actions-show');
  });
});
$(function () {

    'use strict';

    var $body = $('body'),
        Slideout,
        BottomSheets,

        CLASS_IS_SELECTED = 'is-selected',

        hasSlideoutTheme = $body.hasClass('qor-theme-slideout'),
        isSlideoutOpened = function(){
            return $body.hasClass('qor-slideout-open');
        },
        isBottomsheetsOpened = function(){
            return $body.hasClass('qor-bottomsheets-open');
        };


    $body.qorBottomSheets();
    if (hasSlideoutTheme) {
        $body.qorSlideout();
    }

    Slideout = $body.data('qor.slideout');
    BottomSheets = $body.data('qor.bottomsheets');

    function clearSelectedCss () {
        $('[data-url]').removeClass(CLASS_IS_SELECTED);
    }

    function toggleSelectedCss (ele) {
        $('[data-url]').removeClass(CLASS_IS_SELECTED);
        ele.addClass(CLASS_IS_SELECTED);
    }

    function collectSelectID () {
        var $checked = $('.qor-table tbody').find('.mdl-checkbox__input:checked'),
            IDs = [];

        if (!$checked.size()) {
            return;
        }

        $checked.each(function () {
            IDs.push($(this).closest('tr').data('primary-key'));
        });

        return IDs;
    }

    $(document).on('click.qor.openUrl', '[data-url]', function (e) {
        var $this = $(this),
            isNewButton = $this.hasClass('qor-button--new'),
            isEditButton = $this.hasClass('qor-button--edit'),
            isInTable = $this.is('.qor-table tr[data-url]') || $this.closest('.qor-js-table').length,
            isActionButton = $this.hasClass('qor-action-button') || $this.hasClass('qor-action--button'),
            openData = $this.data(),
            actionData;

        // if clicking item's menu actions
         if ((openData.method && openData.method.toUpperCase() != "GET") || $(e.target).parents(".qor-button--actions").size() || (!$(e.target).data('url') && $(e.target).is('a')) || (isInTable && isBottomsheetsOpened())) {
//         if ($(e.target).closest('.qor-table__actions').length
            || (!$(e.target).data('url') && $(e.target).is('a'))
            || (isInTable && isBottomsheetsOpened())) {
            return;
        }

        if (isActionButton) {
            actionData = collectSelectID();
            openData = $.extend({}, openData, {
                actionData: actionData
            });
        }

        if (!openData.method || openData.method.toUpperCase() == "GET") {
            // Open in BottmSheet: is action button, open type is bottom-sheet
            if (isActionButton || openData.openType == 'bottom-sheet') {
                BottomSheets.open(openData);
                return false;
            }

            // Slideout or New Page: table items, new button, edit button
            if (isInTable || (isNewButton && !isBottomsheetsOpened()) || isEditButton || openData.openType == 'slideout') {
                if (hasSlideoutTheme) {
                    if ($this.hasClass(CLASS_IS_SELECTED)) {
                        Slideout.hide();
                        clearSelectedCss();
                        return false;
                    } else {
                        Slideout.open(openData);
                        toggleSelectedCss($this);
                        return false;
                    }
                } else {
                    window.location = openData('url');
                }
                return;
            }

            // Open in BottmSheet: slideout is opened or openType is Bottom Sheet
            if (isSlideoutOpened() || (isNewButton && isBottomsheetsOpened())) {
                BottomSheets.open(openData);
                return false;
            }

            // Other clicks
            if (hasSlideoutTheme) {
                Slideout.open(openData);
                return false;
            } else {
                BottomSheets.open(openData);
                return false;
            }

            return false;
        }
    });

});
$(function () {

  'use strict';

  var location = window.location;

  $('.qor-search').each(function () {
    var $this = $(this);
    var $input = $this.find('.qor-search__input');
    var $clear = $this.find('.qor-search__clear');
    var isSearched = !!$input.val();

    var emptySearch = function () {
      var search = location.search.replace(new RegExp($input.attr('name') + '\\=?\\w*'), '');
      if (search == '?'){
        location.href = location.href.split('?')[0];
      } else {
        location.search = location.search.replace(new RegExp($input.attr('name') + '\\=?\\w*'), '');
      }
    };

    $this.closest('.qor-page__header').addClass('has-search');
    $('header.mdl-layout__header').addClass('has-search');

    $clear.on('click', function () {
      if ($input.val() || isSearched) {
        emptySearch();
      } else {
        $this.removeClass('is-dirty');
      }
    });
  });
});
$(function () {

  'use strict';

  $('.qor-js-table .qor-table__content').each(function () {
    var $this = $(this);
    var width = $this.width();
    var parentWidth = $this.parent().width();

    if (width >= 180 && width < parentWidth) {
      $this.css('max-width', parentWidth);
    }
  });

});
