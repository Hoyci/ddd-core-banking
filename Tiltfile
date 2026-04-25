docker_compose('docker-compose.yml')

DB_URL = 'postgres://postgres:postgres@localhost:5432/corebanking?sslmode=disable'
RABBIT_URL = 'amqp://guest:guest@localhost:5672/'

local_resource(
    'onboarding',
    serve_cmd='go run ./cmd/onboarding/...',
    serve_env={
        'DATABASE_URL': DB_URL,
        'RABBITMQ_URL': RABBIT_URL,
    },
    deps=['cmd/onboarding', 'internal/onboarding', 'pkg'],
    resource_deps=['postgres', 'rabbitmq'],
)

local_resource(
    'account',
    serve_cmd='go run ./cmd/account/...',
    serve_env={
        'DATABASE_URL': DB_URL,
        'RABBITMQ_URL': RABBIT_URL,
    },
    deps=['cmd/account', 'internal/account', 'pkg'],
    resource_deps=['postgres', 'rabbitmq'],
)
