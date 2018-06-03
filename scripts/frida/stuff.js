var modMap = new ModuleMap();
// look for SigKey in libstrings.so
Interceptor.attach(Module.findExportByName(null, 'strlen'), {
	onEnter: function(args) {
		var retMod = modMap.find(this.returnAddress);
		if(retMod.name == "libstrings.so") {
			console.log(
				"\n---------------------------\n"
				+ "SigKey found: '" +
				Memory.readUtf8String(args[0]) + "'" +
				"\n---------------------------\n"
			)
			// this is heavy stuff, better to detach rigth after signature will be found
			Interceptor.detachAll()
		}
	},
})

// dump http requests
// @TODO: dump responce?
// @TODO: it seems there is two separate branches of requests logic and only one is captured here
Java.perform(function (){
	var HTTPRequestHandler = Java.use("com.facebook.proxygen.HTTPRequestHandler");
	var HttpGet = Java.use("org.apache.http.client.methods.HttpGet");
	var HttpPost = Java.use("org.apache.http.client.methods.HttpPost");
	var ByteArrayOutputStream = Java.use("java.io.ByteArrayOutputStream");
	var ByteArrayEntity = Java.use("org.apache.http.entity.ByteArrayEntity");
	
	HTTPRequestHandler.executeWithDefragmentation.implementation = function(req){
		var data = req.getMethod() + " " + req.getURI();
		var casted;
		var body;
		
		switch(req.$className) {
		case "org.apache.http.client.methods.HttpGet":
			casted = Java.cast(req, HttpGet);
			break;
		
		case "org.apache.http.client.methods.HttpPost":
			casted = Java.cast(req, HttpPost);
			var entity = casted.getEntity();
			var stream = ByteArrayOutputStream.$new();
			entity.writeTo(stream);
			body = stream.toString();
			// set copy of request back on
			var out = ByteArrayEntity.$new(stream.toByteArray());
			out.setContentType(entity.getContentType());
			casted.setEntity(out);
			break;
		
		default:
			console.log(data + "\nunexpected request type " + req.$className + '\n');
			return;
		}
		
		casted.getAllHeaders().forEach(function(item){
			data += "\n" + item.getName() + ": \"" + item.getValue() + '"';
		});
		
		if(body) {
			data += "\n\n" + body;
		}
		console.log(data + "\n\n")
		return this.executeWithDefragmentation(req);
	}
});
