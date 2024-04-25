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
    docker compose --file ${ProjectRoot}/deployments/docker-compose/compose.yaml $args
}

switch ($command) {
    
    "openapi" {
        docker run --rm -ti -v ${ProjectRoot}:/local openapitools/openapi-generator-cli generate -c /local/scripts/generator-cfg.yaml
    }
    "start" {
        try {
            mongo up --detach
            go run ${ProjectRoot}/cmd/pnregistry-api-service
        } finally {
            mongo down
        }
    }
    "mongo" {
        mongo up
    }
    default {
        throw "Unknown command: $command"
    }
}