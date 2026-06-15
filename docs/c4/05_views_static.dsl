// Level 1: System Context
systemContext sms "SystemContext" "The system context diagram for the Server Management System." {
    include *
    autoLayout
}

// Level 2: Container
container sms "Containers" "The container diagram for the Server Management System." {
    include *
    autoLayout
}

// Level 3: Component (Backend API)
component backend "Backend_Components" "The component diagram for the Backend API Container." {
    include identityComp serverComp reportingComp notificationComp frontend db redis elastic smtp
    autoLayout
}

// Level 3: Component (Monitoring Worker)
component monitorWorker "MonitorWorker_Components" "The component diagram for the Monitoring Worker." {
    include monitoringComp redis db elastic externalServers
    autoLayout
}

// Level 3: Component (Daily Scheduler)
component scheduler "Scheduler_Components" "The component diagram for the Daily Scheduler." {
    include schedulerComp db elastic smtp
    autoLayout
}
