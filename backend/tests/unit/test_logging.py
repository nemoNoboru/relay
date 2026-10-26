import json
import logging

from relay.shared.logging import JsonFormatter


def test_formats_record_as_json_with_extras() -> None:
    record = logging.LogRecord("relay.test", logging.INFO, __file__, 1, "hello %s", ("you",), None)
    record.request_id = "abc"
    payload = json.loads(JsonFormatter().format(record))
    assert payload["message"] == "hello you"
    assert payload["level"] == "INFO"
    assert payload["logger"] == "relay.test"
    assert payload["request_id"] == "abc"
