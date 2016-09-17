// init for slideout after show event
$.fn.qorSliderAfterShow = {};

// change Mustache tags from {{}} to [[]]
window.Mustache.tags = ['[[', ']]'];

$(document).ajaxComplete(function( event, xhr, settings ) {
    if (settings.type == "POST" || settings.type == "PUT") {
        $.fn.qorSlideoutBeforeHide = null;
        window.onbeforeunload = null;
    }

});