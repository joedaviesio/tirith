"""Monkey-patching logic for Anthropic and OpenAI SDKs."""

import logging
import urllib.request

from costwatch._config import get_proxy_url

logger = logging.getLogger("tirith")

_patched = False


def _is_proxy_running(proxy_url: str) -> bool:
    """Check if the Tirith proxy is reachable."""
    try:
        req = urllib.request.Request(f"{proxy_url}/health", method="GET")
        with urllib.request.urlopen(req, timeout=1) as resp:
            return resp.status == 200
    except Exception:
        return False


def patch_all() -> None:
    """Patch all available AI SDKs to route through Tirith proxy."""
    global _patched

    proxy_url = get_proxy_url()

    if not _is_proxy_running(proxy_url):
        if not _patched:
            logger.warning(
                "Tirith proxy not running at %s — "
                "API calls will go directly to providers. "
                "Run 'tirith start' to enable cost tracking.",
                proxy_url,
            )
        return

    _patch_anthropic(proxy_url)
    _patch_openai(proxy_url)
    _patched = True


def _patch_anthropic(proxy_url: str) -> None:
    """Patch the Anthropic SDK to route through Tirith."""
    try:
        import anthropic
    except ImportError:
        return

    anthropic_base = f"{proxy_url}/proxy/anthropic"

    # Patch sync client.
    _original_init = anthropic.Anthropic.__init__

    def _patched_init(self, *args, **kwargs):
        kwargs.setdefault("base_url", anthropic_base)
        _original_init(self, *args, **kwargs)

    if not getattr(anthropic.Anthropic.__init__, "_costwatch_patched", False):
        anthropic.Anthropic.__init__ = _patched_init
        anthropic.Anthropic.__init__._costwatch_patched = True
        logger.debug("Patched anthropic.Anthropic to use %s", anthropic_base)

    # Patch async client.
    if hasattr(anthropic, "AsyncAnthropic"):
        _original_async_init = anthropic.AsyncAnthropic.__init__

        def _patched_async_init(self, *args, **kwargs):
            kwargs.setdefault("base_url", anthropic_base)
            _original_async_init(self, *args, **kwargs)

        if not getattr(anthropic.AsyncAnthropic.__init__, "_costwatch_patched", False):
            anthropic.AsyncAnthropic.__init__ = _patched_async_init
            anthropic.AsyncAnthropic.__init__._costwatch_patched = True
            logger.debug("Patched anthropic.AsyncAnthropic to use %s", anthropic_base)


def _patch_openai(proxy_url: str) -> None:
    """Patch the OpenAI SDK to route through Tirith."""
    try:
        import openai
    except ImportError:
        return

    openai_base = f"{proxy_url}/proxy/openai"

    # Patch sync client.
    _original_init = openai.OpenAI.__init__

    def _patched_init(self, *args, **kwargs):
        kwargs.setdefault("base_url", openai_base)
        _original_init(self, *args, **kwargs)

    if not getattr(openai.OpenAI.__init__, "_costwatch_patched", False):
        openai.OpenAI.__init__ = _patched_init
        openai.OpenAI.__init__._costwatch_patched = True
        logger.debug("Patched openai.OpenAI to use %s", openai_base)

    # Patch async client.
    if hasattr(openai, "AsyncOpenAI"):
        _original_async_init = openai.AsyncOpenAI.__init__

        def _patched_async_init(self, *args, **kwargs):
            kwargs.setdefault("base_url", openai_base)
            _original_async_init(self, *args, **kwargs)

        if not getattr(openai.AsyncOpenAI.__init__, "_costwatch_patched", False):
            openai.AsyncOpenAI.__init__ = _patched_async_init
            openai.AsyncOpenAI.__init__._costwatch_patched = True
            logger.debug("Patched openai.AsyncOpenAI to use %s", openai_base)
