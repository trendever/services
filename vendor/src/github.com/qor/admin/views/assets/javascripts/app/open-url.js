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
        if ($(e.target).closest('.qor-table__actions').length || (!$(e.target).data('url') && $(e.target).is('a')) || (isInTable && isBottomsheetsOpened())) {
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
