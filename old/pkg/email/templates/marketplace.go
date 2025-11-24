package templates

import (
	"fmt"
	"html"
	"time"
)

// MarketplaceEmailData contains data for marketplace email templates
type MarketplaceEmailData struct {
	OrderID        uint
	ServiceTitle   string
	BuyerName      string
	SellerName     string
	OrderAmount    string
	DueDate        *time.Time
	AdditionalInfo string
}

// NewOrderReceived creates email for seller when they receive a new order
func NewOrderReceived(data MarketplaceEmailData) (subject, htmlBody, textBody string) {
	subject = "🎉 New Order Received!"

	dueDateHTML := ""
	if data.DueDate != nil && !data.DueDate.IsZero() {
		dueDateHTML = fmt.Sprintf(`
			<div style="margin: 16px 0; padding: 12px; background-color: #fafafa; border-left: 3px solid #000000; border-radius: 4px;">
				<strong style="color: #09090b;">Due Date:</strong> <span style="color: #52525b;">%s</span>
			</div>
		`, data.DueDate.Format("January 2, 2006"))
	}

	content := fmt.Sprintf(`<p style="margin: 0 0 20px 0;">Great news! You've received a new order for your service.</p>

<div style="margin: 20px 0;">
	<strong style="color: #09090b; display: block; margin-bottom: 12px;">Order Details:</strong>
	<ul style="margin: 0; padding-left: 20px; color: #52525b;">
		<li style="margin-bottom: 8px;"><strong>Order ID:</strong> #%d</li>
		<li style="margin-bottom: 8px;"><strong>Service:</strong> %s</li>
		<li style="margin-bottom: 8px;"><strong>Buyer:</strong> %s</li>
	</ul>
</div>
%s
<p style="margin: 20px 0 0 0;">The buyer is eager to work with you. Start working on this order to provide excellent service!</p>`,
		data.OrderID,
		html.EscapeString(data.ServiceTitle),
		html.EscapeString(data.BuyerName),
		dueDateHTML,
	)

	htmlBody = BaseEmailTemplate(
		subject,
		content,
		"View Order Details",
		fmt.Sprintf("https://app.hellobutter.io/marketplace/seller/my-sales/%d", data.OrderID),
	)

	textBody = fmt.Sprintf(`Great news! You've received a new order for your service.

Order Details:
- Order ID: #%d
- Service: %s
- Buyer: %s

Start working on this order to provide excellent service!

View order: https://app.hellobutter.io/marketplace/seller/my-sales/%d`,
		data.OrderID,
		data.ServiceTitle,
		data.BuyerName,
		data.OrderID,
	)

	return
}

// OrderStarted creates email for buyer when seller starts working
func OrderStarted(data MarketplaceEmailData) (subject, htmlBody, textBody string) {
	subject = "🚀 Your Order is In Progress"

	content := fmt.Sprintf(`<p style="margin: 0 0 20px 0;">Good news! The seller has started working on your order.</p>

<div style="margin: 20px 0;">
	<strong style="color: #09090b; display: block; margin-bottom: 12px;">Order Details:</strong>
	<ul style="margin: 0; padding-left: 20px; color: #52525b;">
		<li style="margin-bottom: 8px;"><strong>Order ID:</strong> #%d</li>
		<li style="margin-bottom: 8px;"><strong>Service:</strong> %s</li>
		<li style="margin-bottom: 8px;"><strong>Seller:</strong> %s</li>
	</ul>
</div>

<p style="margin: 20px 0 0 0;">You'll receive a notification once the seller submits their work for your review.</p>`,
		data.OrderID,
		html.EscapeString(data.ServiceTitle),
		html.EscapeString(data.SellerName),
	)

	htmlBody = BaseEmailTemplate(
		subject,
		content,
		"Track Progress",
		fmt.Sprintf("https://app.hellobutter.io/marketplace/my-orders/%d", data.OrderID),
	)

	textBody = fmt.Sprintf(`Good news! The seller has started working on your order.

Order Details:
- Order ID: #%d
- Service: %s
- Seller: %s

View order: https://app.hellobutter.io/marketplace/my-orders/%d`,
		data.OrderID,
		data.ServiceTitle,
		data.SellerName,
		data.OrderID,
	)

	return
}

// DeliverySubmitted creates email for buyer when seller submits work
func DeliverySubmitted(data MarketplaceEmailData) (subject, htmlBody, textBody string) {
	subject = "📦 Delivery Submitted for Review"

	content := fmt.Sprintf(`<p style="margin: 0 0 20px 0;">The seller has submitted their work for your order!</p>

<div style="margin: 20px 0;">
	<strong style="color: #09090b; display: block; margin-bottom: 12px;">Order Details:</strong>
	<ul style="margin: 0; padding-left: 20px; color: #52525b;">
		<li style="margin-bottom: 8px;"><strong>Order ID:</strong> #%d</li>
		<li style="margin-bottom: 8px;"><strong>Service:</strong> %s</li>
		<li style="margin-bottom: 8px;"><strong>Seller:</strong> %s</li>
	</ul>
</div>

<p style="margin: 20px 0 0 0;">Please review the delivered work and provide your feedback. You can accept the delivery or request revisions if needed.</p>`,
		data.OrderID,
		html.EscapeString(data.ServiceTitle),
		html.EscapeString(data.SellerName),
	)

	htmlBody = BaseEmailTemplate(
		subject,
		content,
		"Review Delivery",
		fmt.Sprintf("https://app.hellobutter.io/marketplace/my-orders/%d", data.OrderID),
	)

	textBody = fmt.Sprintf(`The seller has submitted their work for your order!

Order Details:
- Order ID: #%d
- Service: %s
- Seller: %s

Review delivery: https://app.hellobutter.io/marketplace/my-orders/%d`,
		data.OrderID,
		data.ServiceTitle,
		data.SellerName,
		data.OrderID,
	)

	return
}

// DeliveryAccepted creates email for seller when buyer accepts their work
func DeliveryAccepted(data MarketplaceEmailData) (subject, htmlBody, textBody string) {
	subject = "✅ Delivery Accepted!"

	content := fmt.Sprintf(`<p style="margin: 0 0 20px 0;">Congratulations! The buyer has accepted your delivered work.</p>

<div style="margin: 20px 0;">
	<strong style="color: #09090b; display: block; margin-bottom: 12px;">Order Details:</strong>
	<ul style="margin: 0; padding-left: 20px; color: #52525b;">
		<li style="margin-bottom: 8px;"><strong>Order ID:</strong> #%d</li>
		<li style="margin-bottom: 8px;"><strong>Service:</strong> %s</li>
		<li style="margin-bottom: 8px;"><strong>Buyer:</strong> %s</li>
	</ul>
</div>

<p style="margin: 20px 0 0 0;">Great job completing this order! Payment will be processed according to your agreement.</p>`,
		data.OrderID,
		html.EscapeString(data.ServiceTitle),
		html.EscapeString(data.BuyerName),
	)

	htmlBody = BaseEmailTemplate(
		subject,
		content,
		"View Order",
		fmt.Sprintf("https://app.hellobutter.io/marketplace/seller/my-sales/%d", data.OrderID),
	)

	textBody = fmt.Sprintf(`Congratulations! The buyer has accepted your delivered work.

Order Details:
- Order ID: #%d
- Service: %s
- Buyer: %s

View order: https://app.hellobutter.io/marketplace/seller/my-sales/%d`,
		data.OrderID,
		data.ServiceTitle,
		data.BuyerName,
		data.OrderID,
	)

	return
}

// DeliveryRejected creates email for seller when buyer rejects their work
func DeliveryRejected(data MarketplaceEmailData) (subject, htmlBody, textBody string) {
	subject = "🔄 Delivery Needs Revision"

	additionalInfoHTML := ""
	if data.AdditionalInfo != "" {
		additionalInfoHTML = fmt.Sprintf(`
<div style="margin: 20px 0; padding: 12px; background-color: #fafafa; border-left: 3px solid #000000; border-radius: 4px;">
	<strong style="color: #09090b;">Feedback:</strong>
	<p style="margin: 8px 0 0 0; color: #52525b;">%s</p>
</div>`, html.EscapeString(data.AdditionalInfo))
	}

	content := fmt.Sprintf(`<p style="margin: 0 0 20px 0;">The buyer has provided feedback on your delivery and requested changes.</p>

<div style="margin: 20px 0;">
	<strong style="color: #09090b; display: block; margin-bottom: 12px;">Order Details:</strong>
	<ul style="margin: 0; padding-left: 20px; color: #52525b;">
		<li style="margin-bottom: 8px;"><strong>Order ID:</strong> #%d</li>
		<li style="margin-bottom: 8px;"><strong>Service:</strong> %s</li>
		<li style="margin-bottom: 8px;"><strong>Buyer:</strong> %s</li>
	</ul>
</div>
%s
<p style="margin: 20px 0 0 0;">Please review the buyer's feedback carefully and submit a revised version of your work.</p>`,
		data.OrderID,
		html.EscapeString(data.ServiceTitle),
		html.EscapeString(data.BuyerName),
		additionalInfoHTML,
	)

	htmlBody = BaseEmailTemplate(
		subject,
		content,
		"View Feedback",
		fmt.Sprintf("https://app.hellobutter.io/marketplace/seller/my-sales/%d", data.OrderID),
	)

	textBody = fmt.Sprintf(`The buyer has provided feedback on your delivery and requested changes.

Order Details:
- Order ID: #%d
- Service: %s
- Buyer: %s

View feedback: https://app.hellobutter.io/marketplace/seller/my-sales/%d`,
		data.OrderID,
		data.ServiceTitle,
		data.BuyerName,
		data.OrderID,
	)

	return
}

// RevisionRequested creates email for seller when buyer requests revision
func RevisionRequested(data MarketplaceEmailData) (subject, htmlBody, textBody string) {
	subject = "📝 Revision Request Received"

	additionalInfoHTML := ""
	if data.AdditionalInfo != "" {
		additionalInfoHTML = fmt.Sprintf(`
<div style="margin: 20px 0; padding: 12px; background-color: #fafafa; border-left: 3px solid #000000; border-radius: 4px;">
	<strong style="color: #09090b;">Request Details:</strong>
	<p style="margin: 8px 0 0 0; color: #52525b;">%s</p>
</div>`, html.EscapeString(data.AdditionalInfo))
	}

	content := fmt.Sprintf(`<p style="margin: 0 0 20px 0;">The buyer has requested changes to your delivery.</p>

<div style="margin: 20px 0;">
	<strong style="color: #09090b; display: block; margin-bottom: 12px;">Order Details:</strong>
	<ul style="margin: 0; padding-left: 20px; color: #52525b;">
		<li style="margin-bottom: 8px;"><strong>Order ID:</strong> #%d</li>
		<li style="margin-bottom: 8px;"><strong>Service:</strong> %s</li>
		<li style="margin-bottom: 8px;"><strong>Buyer:</strong> %s</li>
	</ul>
</div>
%s
<p style="margin: 20px 0 0 0;">Please review the revision request and respond accordingly. You can accept and work on the changes, or discuss with the buyer if needed.</p>`,
		data.OrderID,
		html.EscapeString(data.ServiceTitle),
		html.EscapeString(data.BuyerName),
		additionalInfoHTML,
	)

	htmlBody = BaseEmailTemplate(
		subject,
		content,
		"View Request",
		fmt.Sprintf("https://app.hellobutter.io/marketplace/seller/my-sales/%d", data.OrderID),
	)

	textBody = fmt.Sprintf(`The buyer has requested changes to your delivery.

Order Details:
- Order ID: #%d
- Service: %s

View request: https://app.hellobutter.io/marketplace/seller/my-sales/%d`,
		data.OrderID,
		data.ServiceTitle,
		data.OrderID,
	)

	return
}

// RevisionAccepted creates email for buyer when seller accepts revision request
func RevisionAccepted(data MarketplaceEmailData) (subject, htmlBody, textBody string) {
	subject = "✅ Revision Request Accepted"

	content := fmt.Sprintf(`<p style="margin: 0 0 20px 0;">Great news! The seller has accepted your revision request and will start working on the changes.</p>

<div style="margin: 20px 0;">
	<strong style="color: #09090b; display: block; margin-bottom: 12px;">Order Details:</strong>
	<ul style="margin: 0; padding-left: 20px; color: #52525b;">
		<li style="margin-bottom: 8px;"><strong>Order ID:</strong> #%d</li>
		<li style="margin-bottom: 8px;"><strong>Service:</strong> %s</li>
		<li style="margin-bottom: 8px;"><strong>Seller:</strong> %s</li>
	</ul>
</div>

<p style="margin: 20px 0 0 0;">You'll receive a notification once the seller submits the updated delivery for your review.</p>`,
		data.OrderID,
		html.EscapeString(data.ServiceTitle),
		html.EscapeString(data.SellerName),
	)

	htmlBody = BaseEmailTemplate(
		subject,
		content,
		"View Order",
		fmt.Sprintf("https://app.hellobutter.io/marketplace/my-orders/%d", data.OrderID),
	)

	textBody = fmt.Sprintf(`Good news! The seller has accepted your revision request.

Order Details:
- Order ID: #%d
- Service: %s
- Seller: %s

View order: https://app.hellobutter.io/marketplace/my-orders/%d`,
		data.OrderID,
		data.ServiceTitle,
		data.SellerName,
		data.OrderID,
	)

	return
}

// OrderCompleted creates email for seller when order is completed
func OrderCompleted(data MarketplaceEmailData) (subject, htmlBody, textBody string) {
	subject = "🎊 Order Completed Successfully!"

	content := fmt.Sprintf(`<p style="margin: 0 0 20px 0;">Excellent work! Your order has been marked as completed.</p>

<div style="margin: 20px 0;">
	<strong style="color: #09090b; display: block; margin-bottom: 12px;">Order Details:</strong>
	<ul style="margin: 0; padding-left: 20px; color: #52525b;">
		<li style="margin-bottom: 8px;"><strong>Order ID:</strong> #%d</li>
		<li style="margin-bottom: 8px;"><strong>Service:</strong> %s</li>
		<li style="margin-bottom: 8px;"><strong>Buyer:</strong> %s</li>
	</ul>
</div>

<p style="margin: 20px 0 0 0;">Thank you for providing great service. Keep up the excellent work!</p>`,
		data.OrderID,
		html.EscapeString(data.ServiceTitle),
		html.EscapeString(data.BuyerName),
	)

	htmlBody = BaseEmailTemplate(
		subject,
		content,
		"View Completed Order",
		fmt.Sprintf("https://app.hellobutter.io/marketplace/seller/my-sales/%d", data.OrderID),
	)

	textBody = fmt.Sprintf(`Excellent work! Your order has been marked as completed.

Order Details:
- Order ID: #%d
- Service: %s

View order: https://app.hellobutter.io/marketplace/seller/my-sales/%d`,
		data.OrderID,
		data.ServiceTitle,
		data.OrderID,
	)

	return
}

// OrderCancelled creates email when order is cancelled
func OrderCancelled(data MarketplaceEmailData, isBuyer bool) (subject, htmlBody, textBody string) {
	subject = "❌ Order Cancelled"

	additionalInfoHTML := ""
	if data.AdditionalInfo != "" {
		additionalInfoHTML = fmt.Sprintf(`
<div style="margin: 20px 0; padding: 12px; background-color: #fafafa; border-left: 3px solid #000000; border-radius: 4px;">
	<strong style="color: #09090b;">Cancellation Reason:</strong>
	<p style="margin: 8px 0 0 0; color: #52525b;">%s</p>
</div>`, html.EscapeString(data.AdditionalInfo))
	}

	content := fmt.Sprintf(`<p style="margin: 0 0 20px 0;">An order has been cancelled.</p>

<div style="margin: 20px 0;">
	<strong style="color: #09090b; display: block; margin-bottom: 12px;">Order Details:</strong>
	<ul style="margin: 0; padding-left: 20px; color: #52525b;">
		<li style="margin-bottom: 8px;"><strong>Order ID:</strong> #%d</li>
		<li style="margin-bottom: 8px;"><strong>Service:</strong> %s</li>
	</ul>
</div>
%s
<p style="margin: 20px 0 0 0;">If you have any questions or concerns, please contact support.</p>`,
		data.OrderID,
		html.EscapeString(data.ServiceTitle),
		additionalInfoHTML,
	)

	orderPath := fmt.Sprintf("https://app.hellobutter.io/marketplace/seller/my-sales/%d", data.OrderID)
	if isBuyer {
		orderPath = fmt.Sprintf("https://app.hellobutter.io/marketplace/my-orders/%d", data.OrderID)
	}

	htmlBody = BaseEmailTemplate(
		subject,
		content,
		"View Order Details",
		orderPath,
	)

	textBody = fmt.Sprintf(`An order has been cancelled.

Order Details:
- Order ID: #%d
- Service: %s

View order: %s`,
		data.OrderID,
		data.ServiceTitle,
		orderPath,
	)

	return
}

// DisputeOpened creates email when dispute is opened
func DisputeOpened(data MarketplaceEmailData) (subject, htmlBody, textBody string) {
	subject = "⚠️ Dispute Opened for Your Order"

	additionalInfoHTML := ""
	if data.AdditionalInfo != "" {
		additionalInfoHTML = fmt.Sprintf(`
<div style="margin: 20px 0; padding: 12px; background-color: #fafafa; border-left: 3px solid #000000; border-radius: 4px;">
	<strong style="color: #09090b;">Dispute Details:</strong>
	<p style="margin: 8px 0 0 0; color: #52525b;">%s</p>
</div>`, html.EscapeString(data.AdditionalInfo))
	}

	content := fmt.Sprintf(`<p style="margin: 0 0 20px 0;">A dispute has been opened for one of your orders.</p>

<div style="margin: 20px 0;">
	<strong style="color: #09090b; display: block; margin-bottom: 12px;">Order Details:</strong>
	<ul style="margin: 0; padding-left: 20px; color: #52525b;">
		<li style="margin-bottom: 8px;"><strong>Order ID:</strong> #%d</li>
		<li style="margin-bottom: 8px;"><strong>Service:</strong> %s</li>
	</ul>
</div>
%s
<p style="margin: 20px 0 0 0;">Our support team will review the case and help resolve the issue. Please provide any relevant information to help with the resolution.</p>`,
		data.OrderID,
		html.EscapeString(data.ServiceTitle),
		additionalInfoHTML,
	)

	htmlBody = BaseEmailTemplate(
		subject,
		content,
		"View Dispute",
		fmt.Sprintf("https://app.hellobutter.io/marketplace/seller/my-sales/%d", data.OrderID),
	)

	textBody = fmt.Sprintf(`A dispute has been opened for one of your orders.

Order Details:
- Order ID: #%d
- Service: %s

View dispute: https://app.hellobutter.io/marketplace/seller/my-sales/%d`,
		data.OrderID,
		data.ServiceTitle,
		data.OrderID,
	)

	return
}

// DisputeResolved creates email when dispute is resolved
func DisputeResolved(data MarketplaceEmailData, isBuyer bool) (subject, htmlBody, textBody string) {
	subject = "✓ Dispute Resolved"

	additionalInfoHTML := ""
	if data.AdditionalInfo != "" {
		additionalInfoHTML = fmt.Sprintf(`
<div style="margin: 20px 0; padding: 12px; background-color: #fafafa; border-left: 3px solid #000000; border-radius: 4px;">
	<strong style="color: #09090b;">Resolution:</strong>
	<p style="margin: 8px 0 0 0; color: #52525b;">%s</p>
</div>`, html.EscapeString(data.AdditionalInfo))
	}

	content := fmt.Sprintf(`<p style="margin: 0 0 20px 0;">The dispute for your order has been resolved.</p>

<div style="margin: 20px 0;">
	<strong style="color: #09090b; display: block; margin-bottom: 12px;">Order Details:</strong>
	<ul style="margin: 0; padding-left: 20px; color: #52525b;">
		<li style="margin-bottom: 8px;"><strong>Order ID:</strong> #%d</li>
		<li style="margin-bottom: 8px;"><strong>Service:</strong> %s</li>
	</ul>
</div>
%s
<p style="margin: 20px 0 0 0;">If you have any questions about the resolution, please contact support.</p>`,
		data.OrderID,
		html.EscapeString(data.ServiceTitle),
		additionalInfoHTML,
	)

	orderPath := fmt.Sprintf("https://app.hellobutter.io/marketplace/seller/my-sales/%d", data.OrderID)
	if isBuyer {
		orderPath = fmt.Sprintf("https://app.hellobutter.io/marketplace/my-orders/%d", data.OrderID)
	}

	htmlBody = BaseEmailTemplate(
		subject,
		content,
		"View Order",
		orderPath,
	)

	textBody = fmt.Sprintf(`The dispute for your order has been resolved.

Order Details:
- Order ID: #%d
- Service: %s

View order: %s`,
		data.OrderID,
		data.ServiceTitle,
		orderPath,
	)

	return
}

// DeadlineExtended creates email when deadline is extended
func DeadlineExtended(data MarketplaceEmailData) (subject, htmlBody, textBody string) {
	subject = "📅 Order Deadline Extended"

	dueDateHTML := ""
	if data.DueDate != nil && !data.DueDate.IsZero() {
		dueDateHTML = fmt.Sprintf(`
<div style="margin: 20px 0; padding: 12px; background-color: #fafafa; border-left: 3px solid #000000; border-radius: 4px;">
	<strong style="color: #09090b;">New Due Date:</strong> <span style="color: #52525b;">%s</span>
</div>`, data.DueDate.Format("January 2, 2006 at 3:04 PM"))
	}

	additionalInfoHTML := ""
	if data.AdditionalInfo != "" {
		additionalInfoHTML = fmt.Sprintf(`
<div style="margin: 20px 0; padding: 12px; background-color: #fafafa; border-left: 3px solid #000000; border-radius: 4px;">
	<strong style="color: #09090b;">Details:</strong>
	<p style="margin: 8px 0 0 0; color: #52525b;">%s</p>
</div>`, html.EscapeString(data.AdditionalInfo))
	}

	content := fmt.Sprintf(`<p style="margin: 0 0 20px 0;">The deadline for your order has been extended.</p>

<div style="margin: 20px 0;">
	<strong style="color: #09090b; display: block; margin-bottom: 12px;">Order Details:</strong>
	<ul style="margin: 0; padding-left: 20px; color: #52525b;">
		<li style="margin-bottom: 8px;"><strong>Order ID:</strong> #%d</li>
		<li style="margin-bottom: 8px;"><strong>Service:</strong> %s</li>
	</ul>
</div>
%s%s`,
		data.OrderID,
		html.EscapeString(data.ServiceTitle),
		additionalInfoHTML,
		dueDateHTML,
	)

	htmlBody = BaseEmailTemplate(
		subject,
		content,
		"View Order",
		fmt.Sprintf("https://app.hellobutter.io/marketplace/my-orders/%d", data.OrderID),
	)

	textBody = fmt.Sprintf(`The deadline for your order has been extended.

Order Details:
- Order ID: #%d
- Service: %s
%s

View order: https://app.hellobutter.io/marketplace/my-orders/%d`,
		data.OrderID,
		data.ServiceTitle,
		func() string {
			if data.DueDate != nil {
				return fmt.Sprintf("\n- New Due Date: %s", data.DueDate.Format("January 2, 2006"))
			}
			return ""
		}(),
		data.OrderID,
	)

	return
}

// OrderPaymentCompleted creates email when payment is completed
func OrderPaymentCompleted(data MarketplaceEmailData) (subject, htmlBody, textBody string) {
	subject = "💳 Payment Completed Successfully"

	amountHTML := ""
	if data.OrderAmount != "" {
		amountHTML = fmt.Sprintf(`<li style="margin-bottom: 8px;"><strong>Amount:</strong> %s</li>
`, html.EscapeString(data.OrderAmount))
	}

	content := fmt.Sprintf(`<p style="margin: 0 0 20px 0;">Your payment has been processed successfully!</p>

<div style="margin: 20px 0;">
	<strong style="color: #09090b; display: block; margin-bottom: 12px;">Order Details:</strong>
	<ul style="margin: 0; padding-left: 20px; color: #52525b;">
		<li style="margin-bottom: 8px;"><strong>Order ID:</strong> #%d</li>
		<li style="margin-bottom: 8px;"><strong>Service:</strong> %s</li>
		%s
	</ul>
</div>

<p style="margin: 20px 0 0 0;">The seller will now start working on your order. You'll receive updates as progress is made.</p>`,
		data.OrderID,
		html.EscapeString(data.ServiceTitle),
		amountHTML,
	)

	htmlBody = BaseEmailTemplate(
		subject,
		content,
		"View Order",
		fmt.Sprintf("https://app.hellobutter.io/marketplace/my-orders/%d", data.OrderID),
	)

	textBody = fmt.Sprintf(`Your payment has been processed successfully!

Order Details:
- Order ID: #%d
- Service: %s

View order: https://app.hellobutter.io/marketplace/my-orders/%d`,
		data.OrderID,
		data.ServiceTitle,
		data.OrderID,
	)

	return
}

// OrderReadyToStart creates email for seller when order is paid and has required info
func OrderReadyToStart(data MarketplaceEmailData) (subject, htmlBody, textBody string) {
	subject = "✅ Order Ready to Start!"

	content := fmt.Sprintf(`<p style="margin: 0 0 20px 0;">Great news! An order is now ready for you to begin working on it.</p>

<div style="margin: 20px 0;">
	<strong style="color: #09090b; display: block; margin-bottom: 12px;">Order Details:</strong>
	<ul style="margin: 0; padding-left: 20px; color: #52525b;">
		<li style="margin-bottom: 8px;"><strong>Order ID:</strong> #%d</li>
		<li style="margin-bottom: 8px;"><strong>Service:</strong> %s</li>
		<li style="margin-bottom: 8px;"><strong>Buyer:</strong> %s</li>
	</ul>
</div>

<p style="margin: 20px 0 0 0;">The payment has been completed and all required information has been provided by the buyer. You can now start working on this order!</p>`,
		data.OrderID,
		html.EscapeString(data.ServiceTitle),
		html.EscapeString(data.BuyerName),
	)

	htmlBody = BaseEmailTemplate(
		subject,
		content,
		"Start Working on Order",
		fmt.Sprintf("https://app.hellobutter.io/marketplace/seller/my-sales/%d", data.OrderID),
	)

	textBody = fmt.Sprintf(`Great news! An order is now ready for you to begin working on it.

Order Details:
- Order ID: #%d
- Service: %s
- Buyer: %s

The payment has been completed and all required information has been provided. You can now start working on this order!

View order: https://app.hellobutter.io/marketplace/seller/my-sales/%d`,
		data.OrderID,
		data.ServiceTitle,
		data.BuyerName,
		data.OrderID,
	)

	return
}
