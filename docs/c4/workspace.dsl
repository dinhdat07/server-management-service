workspace "Server Management System" "Modular Architecture documentation for SMS using C4 model." {

    model {
        // Actors & External Systems
        !include 01_systems.dsl
        
        // Target System (sms)
        sms = softwareSystem "Server Management System" "Provides server monitoring, management, and reporting." {
            // Containers
            !include 02_containers.dsl
        }


        // Relationships
        !include 04_relationships.dsl
    }

    views {
        !include 05_views_static.dsl
        !include 06_views_dynamic.dsl
        !include 07_styles.dsl
    }
}
