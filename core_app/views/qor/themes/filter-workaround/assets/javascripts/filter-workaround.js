function onFilterChangeValue(filter) {
	data = decodeSearch(location.search);
	var old_value;
	data = data.filter(function(val) {
		console.log(val, val.startsWith(filter.name + '='));
		if (val.startsWith(filter.name + '=')) {
			old_value = val.slice(filter.name.length + 1);
			return false;
		}
		return true;
	});
	if (filter.value == old_value) {
		return;
	}
	if (filter.value) {
		data.push(filter.name + '=' + filter.value)
	}
	window.location.search = '?' + data.join('&');
}

// copy-paste from qor code
function decodeSearch(search) {
	var data = [];
	if (search && search.indexOf('?') > -1) {
		search = search.split('?')[1];
		if (search && search.indexOf('#') > -1) {
			search = search.split('#')[0];
		}
		if (search) {
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
