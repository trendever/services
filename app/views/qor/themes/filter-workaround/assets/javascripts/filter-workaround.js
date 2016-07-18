function onFilterChangeValue(filter) {

  console.log(filter.name, filter.value)

  if ($.query.get(filter.name) != filter.value) {
    window.location.href = $.query.set(filter.name, filter.value).toString();
  }
}
