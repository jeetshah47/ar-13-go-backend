package email

import (
	"fmt"

	"github.com/ar-13-go-backend/internal/models"
)

// getNotificationEmailTemplate returns the HTML template for notification emails
func getNotificationEmailTemplate(notification *models.Notification) string {
	notificationIcons := map[models.NotificationType]string{
		models.NotificationTypeProjectCreated:       "📁",
		models.NotificationTypeTaskCreated:          "➕",
		models.NotificationTypeTaskAssigned:         "📋",
		models.NotificationTypeProjectUpdated:       "🔄",
		models.NotificationTypeTaskUpdated:          "✏️",
		models.NotificationTypeLeaveRequestCreated:  "📝",
		models.NotificationTypeLeaveRequestApproved: "✅",
		models.NotificationTypeLeaveRequestRejected: "❌",
		models.NotificationTypeUserLogin:            "🔓",
		models.NotificationTypeUserLogout:           "🔒",
	}

	icon := notificationIcons[notification.Type]
	if icon == "" {
		icon = "🔔"
	}

	return fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
      <meta charset="utf-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <style>
        body {
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI',
            Roboto, 'Helvetica Neue', Arial, sans-serif;
          line-height: 1.6;
          color: #333;
          max-width: 600px;
          margin: 0 auto;
          padding: 20px;
          background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
        }
        .container {
          background-color: #ffffff;
          border-radius: 12px;
          padding: 0;
          box-shadow: 0 10px 40px rgba(0,0,0,0.15);
          overflow: hidden;
        }
        .header-banner {
          background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
          padding: 40px 30px;
          text-align: center;
          position: relative;
        }
        .header-banner::before {
          content: '';
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="20" cy="20" r="2" fill="rgba(255,255,255,0.1)"/><circle cx="80" cy="40" r="1.5" fill="rgba(255,255,255,0.1)"/><circle cx="40" cy="70" r="1" fill="rgba(255,255,255,0.1)"/><circle cx="90" cy="80" r="2" fill="rgba(255,255,255,0.1)"/></svg>');
          opacity: 0.3;
        }
        .icon-wrapper {
          position: relative;
          display: inline-block;
          width: 80px;
          height: 80px;
          background: rgba(255, 255, 255, 0.2);
          border-radius: 50%%;
          margin-bottom: 20px;
          display: flex;
          align-items: center;
          justify-content: center;
          backdrop-filter: blur(10px);
          border: 2px solid rgba(255, 255, 255, 0.3);
        }
        .icon {
          font-size: 40px;
          filter: drop-shadow(0 2px 4px rgba(0,0,0,0.2));
        }
        .title {
          font-size: 28px;
          font-weight: 700;
          color: #ffffff;
          margin: 0;
          text-shadow: 0 2px 4px rgba(0,0,0,0.2);
        }
        .content {
          padding: 40px 30px;
        }
        .message {
          font-size: 16px;
          color: #555;
          margin-bottom: 30px;
          padding: 25px;
          background: linear-gradient(135deg, #f8f9fa 0%%, #e9ecef 100%%);
          border-radius: 10px;
          border-left: 4px solid #667eea;
          line-height: 1.8;
        }
        .message p {
          margin: 0 0 10px 0;
        }
        .message p:last-child {
          margin-bottom: 0;
        }
        .footer {
          margin-top: 30px;
          padding: 25px 30px;
          background-color: #f8f9fa;
          font-size: 13px;
          color: #888;
          text-align: center;
          border-top: 1px solid #e0e0e0;
        }
        .divider {
          height: 1px;
          background: linear-gradient(90deg, transparent, #e0e0e0 50%%, transparent);
          margin: 30px 0;
        }
      </style>
    </head>
    <body>
      <div class="container">
        <div class="header-banner">
          <div class="icon-wrapper">
            <div class="icon">%s</div>
          </div>
          <div class="title">%s</div>
        </div>
        <div class="content">
          <div class="message">
            %s
          </div>
          <div class="divider"></div>
          <div class="footer">
            <p style="margin: 0;">This is an automated notification. Please do not reply to this email.</p>
          </div>
        </div>
      </div>
    </body>
    </html>
  `, icon, notification.Title, notification.Message)
}

// getAlertEmailTemplate returns the HTML template for alert emails
func getAlertEmailTemplate(title, message, severity string) string {
	alertColors := map[string]string{
		"info":    "#17a2b8",
		"warning": "#ffc107",
		"error":   "#dc3545",
	}

	alertIcons := map[string]string{
		"info":    "ℹ️",
		"warning": "⚠️",
		"error":   "🚨",
	}

	color := alertColors[severity]
	if color == "" {
		color = alertColors["info"]
	}

	icon := alertIcons[severity]
	if icon == "" {
		icon = alertIcons["info"]
	}

	// Define gradient backgrounds for each severity
	gradientMap := map[string]string{
		"info":    "linear-gradient(135deg, #17a2b8 0%%, #138496 100%%)",
		"warning": "linear-gradient(135deg, #ffc107 0%%, #ff9800 100%%)",
		"error":   "linear-gradient(135deg, #dc3545 0%%, #c82333 100%%)",
	}
	gradient := gradientMap[severity]
	if gradient == "" {
		gradient = gradientMap["info"]
	}

	return fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
      <meta charset="utf-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <style>
        body {
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI',
            Roboto, 'Helvetica Neue', Arial, sans-serif;
          line-height: 1.6;
          color: #333;
          max-width: 600px;
          margin: 0 auto;
          padding: 20px;
          background: linear-gradient(135deg, #f5f7fa 0%%, #c3cfe2 100%%);
        }
        .container {
          background-color: #ffffff;
          border-radius: 12px;
          padding: 0;
          box-shadow: 0 10px 40px rgba(0,0,0,0.15);
          overflow: hidden;
          border-left: 5px solid %s;
        }
        .header-banner {
          background: %s;
          padding: 40px 30px;
          text-align: center;
          position: relative;
        }
        .header-banner::after {
          content: '';
          position: absolute;
          bottom: 0;
          left: 0;
          right: 0;
          height: 3px;
          background: %s;
          opacity: 0.3;
        }
        .icon-wrapper {
          position: relative;
          display: inline-block;
          width: 70px;
          height: 70px;
          background: rgba(255, 255, 255, 0.25);
          border-radius: 50%%;
          margin-bottom: 20px;
          display: flex;
          align-items: center;
          justify-content: center;
          backdrop-filter: blur(10px);
          border: 2px solid rgba(255, 255, 255, 0.4);
          box-shadow: 0 4px 15px rgba(0,0,0,0.2);
        }
        .icon {
          font-size: 36px;
          filter: drop-shadow(0 2px 4px rgba(0,0,0,0.3));
        }
        .title {
          font-size: 26px;
          font-weight: 700;
          color: #ffffff;
          margin: 0;
          text-shadow: 0 2px 4px rgba(0,0,0,0.2);
        }
        .content {
          padding: 40px 30px;
        }
        .message {
          font-size: 16px;
          color: #555;
          margin-bottom: 30px;
          padding: 25px;
          background: linear-gradient(135deg, #f8f9fa 0%%, #e9ecef 100%%);
          border-radius: 10px;
          border-left: 4px solid %s;
          line-height: 1.8;
        }
        .message p {
          margin: 0 0 10px 0;
        }
        .message p:last-child {
          margin-bottom: 0;
        }
        .footer {
          margin-top: 30px;
          padding: 25px 30px;
          background-color: #f8f9fa;
          font-size: 13px;
          color: #888;
          text-align: center;
          border-top: 1px solid #e0e0e0;
        }
        .divider {
          height: 1px;
          background: linear-gradient(90deg, transparent, #e0e0e0 50%%, transparent);
          margin: 30px 0;
        }
      </style>
    </head>
    <body>
      <div class="container">
        <div class="header-banner">
          <div class="icon-wrapper">
            <div class="icon">%s</div>
          </div>
          <div class="title">%s</div>
        </div>
        <div class="content">
          <div class="message">
            %s
          </div>
          <div class="divider"></div>
          <div class="footer">
            <p style="margin: 0;">This is an automated alert. Please do not reply to this email.</p>
          </div>
        </div>
      </div>
    </body>
    </html>
  `, color, gradient, color, color, icon, title, message)
}

// getSignupEmailTemplate returns the HTML template for signup emails
func getSignupEmailTemplate(data SignupEmailData) string {
	return fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
      <meta charset="utf-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <style>
        body {
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI',
            Roboto, 'Helvetica Neue', Arial, sans-serif;
          line-height: 1.6;
          color: #333;
          max-width: 600px;
          margin: 0 auto;
          padding: 20px;
          background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
        }
        .container {
          background-color: #ffffff;
          border-radius: 12px;
          padding: 0;
          box-shadow: 0 10px 40px rgba(0,0,0,0.2);
          overflow: hidden;
        }
        .header-banner {
          background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
          padding: 50px 30px;
          text-align: center;
          position: relative;
        }
        .header-banner::before {
          content: '';
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200"><circle cx="30" cy="30" r="3" fill="rgba(255,255,255,0.15)"/><circle cx="170" cy="50" r="2" fill="rgba(255,255,255,0.15)"/><circle cx="60" cy="120" r="2.5" fill="rgba(255,255,255,0.15)"/><circle cx="180" cy="150" r="3" fill="rgba(255,255,255,0.15)"/><circle cx="100" cy="80" r="1.5" fill="rgba(255,255,255,0.15)"/></svg>');
          opacity: 0.4;
        }
        .icon-wrapper {
          position: relative;
          display: inline-block;
          width: 100px;
          height: 100px;
          background: rgba(255, 255, 255, 0.25);
          border-radius: 50%%;
          margin-bottom: 25px;
          display: flex;
          align-items: center;
          justify-content: center;
          backdrop-filter: blur(10px);
          border: 3px solid rgba(255, 255, 255, 0.4);
          box-shadow: 0 8px 25px rgba(0,0,0,0.2);
        }
        .icon-wrapper::after {
          content: '';
          position: absolute;
          width: 120px;
          height: 120px;
          border: 2px solid rgba(255, 255, 255, 0.2);
          border-radius: 50%%;
          animation: pulse 2s ease-in-out infinite;
        }
        @keyframes pulse {
          0%%, 100%% { transform: scale(1); opacity: 1; }
          50%% { transform: scale(1.1); opacity: 0.7; }
        }
        .icon {
          font-size: 50px;
          filter: drop-shadow(0 2px 6px rgba(0,0,0,0.3));
          z-index: 1;
        }
        .title {
          font-size: 32px;
          font-weight: 700;
          color: #ffffff;
          margin: 0;
          text-shadow: 0 2px 6px rgba(0,0,0,0.3);
          letter-spacing: -0.5px;
        }
        .subtitle {
          font-size: 16px;
          color: rgba(255, 255, 255, 0.9);
          margin-top: 10px;
          font-weight: 300;
        }
        .content {
          padding: 40px 30px;
        }
        .message {
          font-size: 16px;
          color: #555;
          margin-bottom: 30px;
          padding: 30px;
          background: linear-gradient(135deg, #f8f9fa 0%%, #e9ecef 100%%);
          border-radius: 12px;
          line-height: 1.8;
        }
        .message p {
          margin: 0 0 15px 0;
        }
        .message p:last-child {
          margin-bottom: 0;
        }
        .greeting {
          font-size: 18px;
          color: #2c3e50;
          margin-bottom: 20px;
        }
        .button-container {
          text-align: center;
          margin: 30px 0;
        }
        .button {
          display: inline-block;
          padding: 16px 40px;
          background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
          color: #ffffff !important;
          text-decoration: none;
          border-radius: 30px;
          font-weight: 600;
          font-size: 16px;
          box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
          transition: all 0.3s ease;
          letter-spacing: 0.5px;
        }
        .button:hover {
          transform: translateY(-2px);
          box-shadow: 0 6px 20px rgba(102, 126, 234, 0.5);
        }
        .link-section {
          margin: 25px 0;
          padding: 20px;
          background-color: #ffffff;
          border-radius: 8px;
          border: 1px dashed #dee2e6;
        }
        .link-label {
          font-size: 13px;
          color: #6c757d;
          margin-bottom: 8px;
          font-weight: 500;
        }
        .link {
          word-break: break-all;
          color: #667eea;
          text-decoration: none;
          font-size: 14px;
          font-family: 'Courier New', monospace;
        }
        .link:hover {
          text-decoration: underline;
        }
        .warning {
          background: linear-gradient(135deg, #fff3cd 0%%, #ffe69c 100%%);
          border: 2px solid #ffc107;
          border-radius: 10px;
          padding: 20px;
          margin: 25px 0;
          font-size: 14px;
          box-shadow: 0 2px 8px rgba(255, 193, 7, 0.2);
        }
        .warning strong {
          color: #856404;
          display: block;
          margin-bottom: 5px;
        }
        .footer {
          margin-top: 30px;
          padding: 25px 30px;
          background-color: #f8f9fa;
          font-size: 13px;
          color: #888;
          text-align: center;
          border-top: 1px solid #e0e0e0;
        }
        .divider {
          height: 1px;
          background: linear-gradient(90deg, transparent, #e0e0e0 50%%, transparent);
          margin: 30px 0;
        }
        .sparkle {
          display: inline-block;
          font-size: 20px;
          margin: 0 5px;
          animation: sparkle 1.5s ease-in-out infinite;
        }
        @keyframes sparkle {
          0%%, 100%% { opacity: 1; transform: scale(1); }
          50%% { opacity: 0.5; transform: scale(1.2); }
        }
      </style>
    </head>
    <body>
      <div class="container">
        <div class="header-banner">
          <div class="icon-wrapper">
            <div class="icon">👋</div>
          </div>
          <div class="title">Welcome to the Team!</div>
          <div class="subtitle">We're excited to have you join us</div>
        </div>
        <div class="content">
          <div class="message">
            <p class="greeting">Hello <strong>%s</strong>,</p>
            <p>You have been invited to join our platform. We're thrilled to have you on board! <span class="sparkle">✨</span></p>
            <p>Click the button below to complete your account setup and get started:</p>
            <div class="button-container">
              <a href="%s" class="button">Complete Registration</a>
            </div>
            <div class="link-section">
              <div class="link-label">Or copy and paste this link into your browser:</div>
              <a href="%s" class="link">%s</a>
            </div>
            <div class="warning">
              <strong>⚠️ Important:</strong> This invitation link is valid for 7 days. Please complete your registration as soon as possible to secure your account.
            </div>
          </div>
          <div class="divider"></div>
          <div class="footer">
            <p style="margin: 0;">This invitation was sent by an administrator. If you didn't expect this email, please contact support.</p>
          </div>
        </div>
      </div>
    </body>
    </html>
  `, data.UserName, data.SignupLink, data.SignupLink, data.SignupLink)
}
