package templates

import (
	"fmt"
	"html"
)

// BaseEmailTemplate creates a beautiful HTML email template with consistent styling
// Note: content parameter should already be properly formatted HTML (not escaped)
func BaseEmailTemplate(title, content, ctaText, ctaLink string) string {
	// Only escape the title for safety
	safeTitle := html.EscapeString(title)

	// Build CTA button HTML if provided
	ctaHTML := ""
	if ctaText != "" && ctaLink != "" {
		ctaHTML = fmt.Sprintf(`
			<table cellpadding="0" cellspacing="0" border="0" style="margin: 30px 0;">
				<tr>
					<td style="background-color: #000000; border-radius: 6px; padding: 12px 24px; border: 1px solid #e5e7eb;">
						<a href="%s" style="color: #ffffff; text-decoration: none; font-weight: 600; font-size: 16px; display: inline-block;">%s</a>
					</td>
				</tr>
			</table>
		`, html.EscapeString(ctaLink), html.EscapeString(ctaText))
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>%s</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #fafafa;">
	<table cellpadding="0" cellspacing="0" border="0" width="100%%" style="background-color: #fafafa; padding: 40px 20px;">
		<tr>
			<td align="center">
				<table cellpadding="0" cellspacing="0" border="0" width="600" style="max-width: 600px; background-color: #ffffff; border-radius: 8px; border: 1px solid #e5e7eb;">
					<!-- Header -->
					<tr>
						<td style="background-color: #000000; padding: 32px; border-radius: 8px 8px 0 0; text-align: center;">
							<h1 style="margin: 0; color: #ffffff; font-size: 28px; font-weight: bold; letter-spacing: -0.5px;">🧈 Butter</h1>
						</td>
					</tr>
					<!-- Content -->
					<tr>
						<td style="padding: 40px 32px;">
							<h2 style="margin: 0 0 24px 0; color: #09090b; font-size: 24px; font-weight: 600; letter-spacing: -0.5px;">%s</h2>
							<div style="margin: 0; color: #52525b; font-size: 16px; line-height: 1.6;">%s</div>
							%s
						</td>
					</tr>
					<!-- Footer -->
					<tr>
						<td style="padding: 24px 32px; background-color: #fafafa; border-radius: 0 0 8px 8px; border-top: 1px solid #e5e7eb;">
							<p style="margin: 0 0 8px 0; color: #71717a; font-size: 14px; text-align: center;">
								<a href="https://app.hellobutter.io" style="color: #000000; text-decoration: none; font-weight: 500;">Visit Dashboard</a> · 
								<a href="https://app.hellobutter.io/support" style="color: #000000; text-decoration: none; font-weight: 500;">Get Help</a>
							</p>
							<p style="margin: 0; color: #a1a1aa; font-size: 12px; text-align: center;">
								© 2025 Butter. All rights reserved.
							</p>
						</td>
					</tr>
				</table>
			</td>
		</tr>
	</table>
</body>
</html>
`, safeTitle, safeTitle, content, ctaHTML)
}
