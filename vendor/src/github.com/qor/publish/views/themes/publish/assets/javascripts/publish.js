!function(t){"function"==typeof define&&define.amd?define(["jquery"],t):t("object"==typeof exports?require("jquery"):jQuery)}(function(t){"use strict";function e(e,i){return"string"==typeof e&&"object"==typeof i&&t.each(i,function(t,i){e=e.replace("${"+String(t).toLowerCase()+"}",i)}),e}function i(e,o){this.$element=t(e),this.options=t.extend(!0,{},i.DEFAULTS,t.isPlainObject(o)&&o),this.loading=!1,this.init()}var o="qor.publish",l="click."+o,d=".qor-table",s=".qor-actions--publish";i.prototype={constructor:i,init:function(){this.$modal=t(e(i.MODAL,this.options.text)).appendTo("body"),t(d).size()?this.bind():this.disableButtons()},bind:function(){this.$element.on(l,t.proxy(this.click,this))},unbind:function(){this.$element.off(l,this.click)},disableButtons:function(){t(s).find("button").prop("disabled",!0)},click:function(e){var o,l,d=this.options,s=t(e.target),a=t(".qor-js-table").find("input:checkbox").not(d.toggleCheck).is(":checked");if((s.is(d.submit)||s.is(d.schedulePopoverButton))&&!a)return window.alert(s.data().noItem),!1;if(s.is(d.scheduleSetButton)&&(l=t(d.scheduleTime).val(),l&&(t(".publish-schedule-time").val(l),a?t(d.submit).closest("form").submit():(this.$scheduleModal.qorModal("hide"),this.$scheduleModal.trigger("disable")))),s.is(d.schedulePopoverButton)&&(o=s.data(),this.$scheduleModal&&(this.$scheduleModal.trigger("disable"),this.$scheduleModal.remove()),this.$scheduleModal=t(window.Mustache.render(i.SCHEDULE,o)).appendTo("body"),this.$scheduleModal.qorModal("show"),window.componentHandler.upgradeElement(document.querySelector(".qor-publish__time")),this.$scheduleModal.trigger("enable")),s.is(d.toggleView)){if(this.loading)return;this.loading=!0,this.$modal.find(".mdl-card__supporting-text").empty().load(s.data("url"),t.proxy(this.show,this))}else s.is(d.toggleCheck)&&(s.prop("disabled")||s.closest("table").find("tbody :checkbox").click())},show:function(){this.loading=!1,this.$modal.qorModal("show")},destroy:function(){this.unbind(),this.$element.removeData(o)}},i.DEFAULTS={toggleView:".qor-js-view",toggleCheck:".qor-js-check-all",schedulePopoverButton:".qor-publish__button-popover",scheduleSetButton:".qor-publish__button-schedule",scheduleTime:".qor-publish__time",submit:".qor-publish__submit",text:{title:"Changes",close:"Close"}},i.SCHEDULE='<div class="qor-modal qor-modal-mini fade" tabindex="-1" role="dialog" aria-hidden="true"><div class="mdl-card mdl-shadow--4dp" role="document"><div class="mdl-card__title"><h2 class="mdl-card__title-text">[[modalTitle]]</h2></div><div class="mdl-card__supporting-text"><p class="hint">[[modalHint]]</p><div class="qor-field__datetimepicker qor-publish__datetimepicker"><div class="mdl-textfield mdl-js-textfield"><input class="mdl-textfield__input qor-publish__time ignore-dirtyform" id="qorPublishTime" type="text" placeholder="YYYY-MM-DD HH:MM" data-start-date="true" /><label class="mdl-textfield__label" for="qorPublishTime"></label></div><div><button data-toggle="qor.datepicker" data-target-input=".qor-publish__time" class="mdl-button mdl-js-button mdl-button--icon qor-action__datepicker" type="button"><i class="material-icons">date_range</i></button><button data-toggle="qor.timepicker" data-target-input=".qor-publish__time" class="mdl-button mdl-js-button mdl-button--icon qor-action__timepicker" type="button"><i class="material-icons">access_time</i></button></div></div><div class="mdl-card__actions"><a class="mdl-button mdl-button--colored mdl-js-button qor-publish__button-schedule">[[modalSet]]</a><a class="mdl-button mdl-js-button" data-dismiss="modal">[[modalCancel]]</a></div></div></div></div>',i.MODAL='<div class="qor-modal fade" tabindex="-1" role="dialog" aria-hidden="true"><div class="mdl-card mdl-shadow--4dp" role="document"><div class="mdl-card__title"><h2 class="mdl-card__title-text">${title}</h2></div><div class="mdl-card__supporting-text"></div><div class="mdl-card__actions"><a class="mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect" data-dismiss="modal">${close}</a></div><div class="mdl-card__menu"><button class="mdl-button mdl-button--icon mdl-js-button mdl-js-ripple-effect" data-dismiss="modal" aria-label="close"><i class="material-icons">close</i></button></div></div></div>',i.plugin=function(e){return this.each(function(){var l,d,s=t(this),a=s.data(o);if(!a){if(/destroy/.test(e))return;l=t.extend(!0,{text:s.data("text")},"object"==typeof e&&e),s.data(o,a=new i(this,l))}"string"==typeof e&&t.isFunction(d=a[e])&&d.apply(a)})},t(function(){i.plugin.call(t(".qor-theme-publish"))})});
//# sourceMappingURL=publish.js.map
