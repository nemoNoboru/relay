from relay.shared.domain.models import ValueObject


class HealthStatus(ValueObject):
    status: str
    version: str
