"""Tirith — AI API cost observability.

Import this module to automatically route all Anthropic and OpenAI API calls
through the Tirith proxy for cost tracking.

Usage:
    import tirith  # that's it — patches SDKs automatically

    from anthropic import Anthropic
    client = Anthropic()  # routes through Tirith proxy
"""

from tirith._patch import patch_all as _patch_all
from tirith._config import configure, get_proxy_url

__version__ = "2.0.0"

# Auto-patch on import.
_patch_all()
