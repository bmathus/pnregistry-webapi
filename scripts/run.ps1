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
    docker compose --file ${ProjectRoot}/build/docker-compose/compose.yaml $args
}

switch ($command) {
    
    "start" {
        try {
            mongo up --detach
            go run ${ProjectRoot}/cmd/pnregistry-api-service
        } finally {
            mongo down
        }
    }
    "docker" {
       docker build -t thamako3/pnregistry-webapi:local-build -f ${ProjectRoot}/build/docker/Dockerfile .
    }
    "mongo" {
        mongo up
    }
    default {
        throw "Unknown command: $command"
    }
}