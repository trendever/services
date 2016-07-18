package notifier

//Notifier is an interface for sending notifications
type Notifier interface {
	//NotifyByEmail sends notification to user by email
	//about - template name
	//model - model
	NotifyByEmail(about string, model interface{}) error
	//NotifyBySms sends notification to user by sms
	//about - template name
	//model - model
	NotifyBySms(about string, model interface{}) error
	//NotifyByTelegram sends notification to telegram
	NotifyByTelegram(channel string, message interface{}) error
}

//NotifyBy is a function for sending notification
type NotifyBy func(about string, model interface{}) error

//CallSupplierToChat calls the supplier to the chat
func CallSupplierToChat(supplier, url, lead interface{}, f NotifyBy) error {
	return f("call_supplier_to_chat", struct {
		Supplier interface{}
		URL      interface{}
		Lead     interface{}
	}{
		Supplier: supplier,
		URL:      url,
		Lead:     lead,
	})
}

//CallCustomerToChat calls the customer to the chat
func CallCustomerToChat(customer, url, lead interface{}, f NotifyBy) error {
	return f("call_customer_to_chat", struct {
		Customer interface{}
		URL      interface{}
		Lead     interface{}
	}{
		Customer: customer,
		URL:      url,
		Lead:     lead,
	})
}

//NotifySellerAboutLead notifies the seller about the lead
func NotifySellerAboutLead(seller, url, lead interface{}, f NotifyBy) error {
	return f("notify_seller_about_lead", struct {
		Seller interface{}
		URL    interface{}
		Lead   interface{}
	}{
		Seller: seller,
		URL:    url,
		Lead:   lead,
	})
}

//NotifyCustomerAboutLead sends notification to user
func NotifyCustomerAboutLead(customer, url, lead interface{}, f NotifyBy) error {
	return f("notify_customer_about_lead", struct {
		Customer interface{}
		URL      interface{}
		Lead     interface{}
	}{
		Customer: customer,
		URL:      url,
		Lead:     lead,
	})
}

//NotifySellerAboutUnreadMessage sends notification to seller
func NotifySellerAboutUnreadMessage(seller, url, lead interface{}, f NotifyBy) error {
	return f("notify_seller_about_unread_message", struct {
		Seller interface{}
		URL    interface{}
		Lead   interface{}
	}{
		Seller: seller,
		URL:    url,
		Lead:   lead,
	})
}
