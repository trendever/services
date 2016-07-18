package soso

func (r *Router) CREATE(data_type string, handler HandlerFunc) {
	r.Handle("create", data_type, handler)
}

func (r *Router) RETRIEVE(data_type string, handler HandlerFunc) {
	r.Handle("retrieve", data_type, handler)
}

func (r *Router) UPDATE(data_type string, handler HandlerFunc) {
	r.Handle("update", data_type, handler)
}

func (r *Router) DELETE(data_type string, handler HandlerFunc) {
	r.Handle("delete", data_type, handler)
}

func (r *Router) FLUSH(data_type string, handler HandlerFunc) {
	r.Handle("flush", data_type, handler)
}
