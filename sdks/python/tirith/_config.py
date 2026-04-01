"""Configuration for the Tirith SDK wrapper."""

import os
import re
from pathlib import Path

_DEFAULT_PROXY_PORT = 5555
_proxy_url: str | None = None
_api_key: str | None = None
_default_tags: dict[str, str] = {}


def get_proxy_url() -> str:
    """Get the proxy URL, checking env var, config file, then default."""
    global _proxy_url
    if _proxy_url is not None:
        return _proxy_url

    # Check env var first.
    env_url = os.environ.get("TIRITH_PROXY_URL")
    if env_url:
        return env_url

    # Try to read port from config file (simple regex, avoids yaml dependency).
    config_path = Path.home() / ".tirith" / "config.yaml"
    if config_path.exists():
        try:
            content = config_path.read_text()
            match = re.search(r"^\s*port:\s*(\d+)", content, re.MULTILINE)
            if match:
                return f"http://localhost:{match.group(1)}"
        except Exception:
            pass

    return f"http://localhost:{_DEFAULT_PROXY_PORT}"


def configure(
    proxy_url: str | None = None,
    api_key: str | None = None,
    default_tags: dict[str, str] | None = None,
) -> None:
    """Configure the Tirith SDK wrapper.

    Args:
        proxy_url: Override the proxy URL (e.g., "https://proxy.tirith.dev").
        api_key: API key for cloud mode.
        default_tags: Tags applied to every API call automatically.
    """
    global _proxy_url, _api_key, _default_tags

    if proxy_url is not None:
        _proxy_url = proxy_url
    if api_key is not None:
        _api_key = api_key
    if default_tags is not None:
        _default_tags = default_tags

    # Re-patch with new config.
    from tirith._patch import patch_all
    patch_all()
