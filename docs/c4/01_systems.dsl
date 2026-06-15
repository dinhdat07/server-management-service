// Actors
admin = person "System Administrator" "Manages servers, views reports, and receives notifications."

// External Systems
smtp = softwareSystem "SMTP Server" "External email delivery service (e.g., MailHog/SendGrid)." "External"
externalServers = softwareSystem "Target Servers (10k+)" "The external servers being managed and monitored by SMS." "External"
