[CmdletBinding()]
param (
    [Parameter()]
    [switch]
    $Development
)

if ($Development) {
    docker-compose down
    docker-compose up -d
} else {
    docker-compose -f ./docker-compose.yml -f ./docker-compose.prod.yml down
    docker-compose -f ./docker-compose.yml -f ./docker-compose.prod.yml up -d
}
