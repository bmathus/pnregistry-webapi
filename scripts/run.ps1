param (
    $command
)

if (-not $command)  {
    $command = "start"
}

$ProjectRoot = "${PSScriptRoot}/.."

$env:PN_REGISTRY_API_ENVIRONMENT="Development"
$env:PN_REGISTRY_API_PORT="8080"
$env:PN_REGISTRY_API_MONGODB_USERNAME="root"
$env:PN_REGISTRY_API_MONGODB_PASSWORD="neUhaDnes"

function mongo {
    param (
        $composeFile = "compose.yaml"
    )
    docker compose --file "${ProjectRoot}/build/docker-compose/$composeFile" $args
}


switch ($command) {
    
    "start" {
        try {
            mongo -composeFile "compose.yaml" up --detach
            go run ${ProjectRoot}/cmd/pnregistry-api-service
        } finally {
            mongo -composeFile "compose.yaml" down
        }
    }
    "docker" {
       docker build -t thamako3/pnregistry-webapi:local-build -f ${ProjectRoot}/build/docker/Dockerfile .
    }
    "mongo" {
        mongo -composeFile "compose.yaml" up
    }
    "start-test-db" {
        # Start the test MongoDB + mongo express instance
        mongo -composeFile "compose-test.yaml" up --detach
    }
    "stop-test-db" {
        # Stop the test MongoDB + mongo express instance only
        mongo -composeFile "compose-test.yaml" down
    }
    "test" {
        # Configure test environment variables
        $env:PN_REGISTRY_API_ENVIRONMENT="Testing"
        $env:PN_REGISTRY_API_PORT="8081"
        $env:PN_REGISTRY_API_MONGODB_HOST="localhost"
        $env:PN_REGISTRY_API_MONGODB_PORT="27018"
        $env:PN_REGISTRY_API_MONGODB_DATABASE="pn-registry-test"

        try {
            # Spin up the test MongoDB instance
            mongo -composeFile "compose-test.yaml" up --detach

            # Run tests with test-specific environment variables
            & go test -v ./internal/integration_tests/...
        } finally {
            # Clean up the test MongoDB instance after tests finish
            mongo -composeFile "compose-test.yaml" down
        }
    }
    default {
        throw "Unknown command: $command"
    }
}